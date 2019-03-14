// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controlplane

import (
	"context"
	"fmt"
	"path/filepath"

	apisaws "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/imagevector"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// Object names
const (
	cloudControllerManagerDeploymentName = "cloud-controller-manager"
	cloudControllerManagerServerName     = "cloud-controller-manager-server"
)

var controlPlaneSecrets = &secrets.Secrets{
	CertificateSecretConfigs: map[string]*secrets.CertificateSecretConfig{
		gardencorev1alpha1.SecretNameCACluster: {
			Name:       gardencorev1alpha1.SecretNameCACluster,
			CommonName: "kubernetes",
			CertType:   secrets.CACert,
		},
	},
	SecretConfigsFunc: func(cas map[string]*secrets.Certificate, clusterName string) []secrets.ConfigInterface {
		return []secrets.ConfigInterface{
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         cloudControllerManagerDeploymentName,
					CommonName:   "system:cloud-controller-manager",
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[gardencorev1alpha1.SecretNameCACluster],
				},
				KubeConfigRequest: &secrets.KubeConfigRequest{
					ClusterName:  clusterName,
					APIServerURL: common.KubeAPIServerDeploymentName,
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:       cloudControllerManagerServerName,
					CommonName: cloudControllerManagerDeploymentName,
					DNSNames:   controlplane.DNSNamesForService(cloudControllerManagerDeploymentName, clusterName),
					CertType:   secrets.ServerCert,
					SigningCA:  cas[gardencorev1alpha1.SecretNameCACluster],
				},
			},
		}
	},
}

var configChart = &chart.Chart{
	Name: "cloud-provider-config",
	Path: filepath.Join(aws.InternalChartsPath, "cloud-provider-config"),
	Objects: []*chart.Object{
		{
			Type: &corev1.ConfigMap{},
			Name: common.CloudProviderConfigName,
		},
	},
}

var ccmChart = &chart.Chart{
	Name:   "cloud-controller-manager",
	Path:   filepath.Join(aws.InternalChartsPath, "cloud-controller-manager"),
	Images: []string{common.HyperkubeImageName},
	ValuesFunc: func(clusterName string, shoot *gardenv1beta1.Shoot, checksums map[string]string) (map[string]interface{}, error) {
		return map[string]interface{}{
			"cloudProvider": "aws",
			"clusterName":   clusterName,

			"kubernetesVersion": shoot.Spec.Kubernetes.Version,
			"podNetwork":        extensionscontroller.GetPodNetwork(shoot),
			"replicas":          extensionscontroller.GetReplicas(shoot, 1),
			"podAnnotations": map[string]interface{}{
				"checksum/secret-cloud-controller-manager":        checksums[cloudControllerManagerDeploymentName],
				"checksum/secret-cloud-controller-manager-server": checksums[cloudControllerManagerServerName],
				"checksum/secret-cloudprovider":                   checksums[common.CloudProviderSecretName],
				"checksum/configmap-cloud-provider-config":        checksums[common.CloudProviderConfigName],
			},
			"configureRoutes": false,
			"environment": []map[string]interface{}{
				{
					"name": "AWS_ACCESS_KEY_ID",
					"valueFrom": map[string]interface{}{
						"secretKeyRef": map[string]interface{}{
							"key":  aws.AccessKeyID,
							"name": common.CloudProviderSecretName,
						},
					},
				},
				{
					"name": "AWS_SECRET_ACCESS_KEY",
					"valueFrom": map[string]interface{}{
						"secretKeyRef": map[string]interface{}{
							"key":  aws.SecretAccessKey,
							"name": common.CloudProviderSecretName,
						},
					},
				},
			},
			"resources": map[string]interface{}{
				"limits": map[string]interface{}{
					"cpu":    "500m",
					"memory": "512Mi",
				},
			},
		}, nil
	},
	Objects: []*chart.Object{
		{Type: &corev1.Service{}, Name: "cloud-controller-manager"},
		{Type: &appsv1.Deployment{}, Name: "cloud-controller-manager"},
	},
}

// NewActuator creates a new Actuator that acts upon and updates the status of ControlPlane resources.
func NewActuator() controlplane.Actuator {
	return &actuator{
		logger: log.Log.WithName("controlplane-controller"),
	}
}

// actuator is an Actuator that acts upon and updates the status of ControlPlane resources.
type actuator struct {
	scheme            *runtime.Scheme
	decoder           runtime.Decoder
	config            *rest.Config
	clientset         kubernetes.Interface
	gardenerClientset gardenerkubernetes.Interface
	chartApplier      gardenerkubernetes.ChartApplier
	client            client.Client
	logger            logr.Logger
}

// InjectScheme injects the given scheme into the actuator.
func (a *actuator) InjectScheme(scheme *runtime.Scheme) error {
	a.scheme = scheme
	a.decoder = serializer.NewCodecFactory(a.scheme).UniversalDecoder()
	return nil
}

// InjectConfig injects the given config into the actuator.
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config

	// Create clientset
	var err error
	a.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "could not create Kubernetes client")
	}

	// Create Gardener clientset
	a.gardenerClientset, err = gardenerkubernetes.NewForConfig(config, client.Options{})
	if err != nil {
		return errors.Wrap(err, "could not create Gardener client")
	}

	// Create chart applier
	a.chartApplier, err = gardenerkubernetes.NewChartApplierForConfig(config)
	if err != nil {
		return errors.Wrap(err, "could not create chart applier")
	}

	return nil
}

// InjectClient injects the given client into the actuator.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

// Reconcile reconciles the given controlplane and cluster, creating or updating the additional Shoot
// control plane components as needed.
func (a *actuator) Reconcile(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) error {
	// Decode providerConfig
	cpConfig := &apisaws.ControlPlaneConfig{}
	if _, _, err := a.decoder.Decode(cp.Spec.ProviderConfig.Raw, nil, cpConfig); err != nil {
		return errors.Wrapf(err, "could not decode providerConfig of controlplane '%s'", objectName(cp))
	}

	// Decode infrastructureProviderStatus
	infraStatus := &apisaws.InfrastructureStatus{}
	if _, _, err := a.decoder.Decode(cp.Spec.InfrastructureProviderStatus.Raw, nil, infraStatus); err != nil {
		return errors.Wrapf(err, "could not decode infrastructureProviderStatus of controlplane '%s'", objectName(cp))
	}

	// Deploy secrets
	a.logger.Info("Deploying secrets", "controlplane", objectName(cp))
	deployedSecrets, err := controlPlaneSecrets.Deploy(a.clientset, a.gardenerClientset, cp.Namespace)
	if err != nil {
		return errors.Wrapf(err, "could not deploy secrets for controlplane '%s'", objectName(cp))
	}

	// Determine subnet ID and zone
	subnetID, zone, err := getSubnetIDAndZone(infraStatus)
	if err != nil {
		return errors.Wrapf(err, "could not determine subnet ID or zone from infrastructureProviderStatus of controlplane '%s'", objectName(cp))
	}

	// Collect additional configuration chart values
	values := map[string]interface{}{
		"cloudProviderConfig": getCloudProviderConfig(infraStatus.VPC.ID, subnetID, zone, cp.Namespace),
	}

	// Apply config chart
	a.logger.Info("Applying configuration chart", "controlplane", objectName(cp), "chart", configChart.Name, "values", values)
	if err := configChart.Apply(ctx, a.gardenerClientset, a.chartApplier, cp.Namespace, cluster.Shoot, nil, nil, values); err != nil {
		return errors.Wrapf(err, "could not apply configuration chart for controlplane '%s'", objectName(cp))
	}

	// Compute all needed checksums
	checksums, err := a.computeChecksums(ctx, deployedSecrets, cp.Namespace)
	if err != nil {
		return err
	}

	// Collect additional CCM chart values
	values = make(map[string]interface{})
	if cpConfig.CloudControllerManager != nil {
		values["featureGates"] = cpConfig.CloudControllerManager.FeatureGates
	}

	// Apply CCM chart
	a.logger.Info("Applying CCM chart", "controlplane", objectName(cp), "chart", ccmChart.Name, "values", values)
	if err := ccmChart.Apply(ctx, a.gardenerClientset, a.chartApplier, cp.Namespace, cluster.Shoot, imagevector.ImageVector(), checksums, values); err != nil {
		return errors.Wrapf(err, "could not apply CCM chart for controlplane '%s'", objectName(cp))
	}

	return nil
}

// Delete reconciles the given controlplane and cluster, deleting the additional Shoot
// control plane components as needed.
func (a *actuator) Delete(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) error {
	// Delete CCM objects
	a.logger.Info("Deleting CCM objects", "controlplane", objectName(cp))
	if err := ccmChart.Delete(ctx, a.client, cp.Namespace); err != nil {
		return errors.Wrapf(err, "could not delete CCM objects for controlplane '%s'", objectName(cp))
	}

	// Delete config objects
	a.logger.Info("Deleting configuration objects", "controlplane", objectName(cp))
	if err := configChart.Delete(ctx, a.client, cp.Namespace); err != nil {
		return errors.Wrapf(err, "could not delete configuration objects for controlplane '%s'", objectName(cp))
	}

	// Delete secrets
	a.logger.Info("Deleting secrets", "controlplane", objectName(cp))
	if err := controlPlaneSecrets.Delete(a.clientset, cp.Namespace); err != nil {
		return errors.Wrapf(err, "could not delete secrets for controlplane '%s'", objectName(cp))
	}

	return nil
}

// computeChecksums computes and returns all needed checksums. This includes the checksums for the given deployed secrets,
// as well as the cloud provider secret and configmap that are fetched from the cluster.
func (a *actuator) computeChecksums(
	ctx context.Context,
	deployedSecrets map[string]*corev1.Secret,
	namespace string,
) (map[string]string, error) {
	// Get cloud provider secret and configmap from cluster
	cpSecret := &corev1.Secret{}
	err := a.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: common.CloudProviderSecretName}, cpSecret)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get secret '%s'", objectName(cpSecret))
	}
	cpConfigMap := &corev1.ConfigMap{}
	err = a.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: common.CloudProviderConfigName}, cpConfigMap)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get configmap '%s'", objectName(cpConfigMap))
	}

	// Compute checksums
	csSecrets := controlplane.MergeSecretMaps(deployedSecrets, map[string]*corev1.Secret{
		common.CloudProviderSecretName: cpSecret,
	})
	csConfigMaps := map[string]*corev1.ConfigMap{
		common.CloudProviderConfigName: cpConfigMap,
	}
	return controlplane.ComputeChecksums(csSecrets, csConfigMaps), nil
}

// getSubnetIDAndZone determines the subnet ID and zone from the given infrastructure status by looking for the first
// subnet with purpose "public".
func getSubnetIDAndZone(infraStatus *apisaws.InfrastructureStatus) (string, string, error) {
	for _, subnet := range infraStatus.VPC.Subnets {
		if subnet.Purpose == apisaws.PurposePublic {
			return subnet.ID, subnet.Zone, nil
		}
	}
	return "", "", errors.Errorf("subnet with purpose 'public' not found")
}

// getCloudProviderConfig builds and returns a AWS config from the given parameters.
func getCloudProviderConfig(vpcID, subnetID, zone, clusterID string) string {
	return fmt.Sprintf(
		`[Global]
VPC=%q
SubnetID=%q
DisableSecurityGroupIngress=true
KubernetesClusterTag=%q
KubernetesClusterID=%q
Zone=%q
`,
		vpcID, subnetID, clusterID, clusterID, zone)
}

func objectName(obj runtime.Object) string {
	k, err := client.ObjectKeyFromObject(obj)
	if err != nil {
		return "/"
	}
	return k.String()
}
