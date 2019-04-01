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

package internal

import (
	"fmt"
	"path/filepath"

	apisaws "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	awsimagevector "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/imagevector"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane"
	"github.com/gardener/gardener-extensions/pkg/util"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
)

// Object names
const (
	cloudControllerManagerDeploymentName = "cloud-controller-manager"
	cloudControllerManagerServerName     = "cloud-controller-manager-server"
)

// Helper is a control plane actuator helper type.
type Helper struct {
	// Secrets is a set of Secrets.
	Secrets controlplane.Secrets
	// ConfigChart is a configuration Chart.
	ConfigChart controlplane.Chart
	// ControlPlaneChart is a control plane Chart.
	ControlPlaneChart controlplane.Chart
	// ImageVector is a function that returns the appropriate ImageVector.
	ImageVectorFunc func() imagevector.ImageVector
	// ConfigChartValuesFunc is a function that collects and returns the configuration Chart values.
	ConfigChartValuesFunc func(runtime.Decoder, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster) (map[string]interface{}, error)
	// ControlPlaneValuesFunc is a function that collects and returns the control plane Chart values.
	ControlPlaneChartValuesFunc func(runtime.Decoder, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster, map[string]string) (map[string]interface{}, error)
}

// NewHelper creates and returns a new Helper instance.
func NewHelper() *Helper {
	return &Helper{
		Secrets:                     controlPlaneSecrets,
		ConfigChart:                 configChart,
		ControlPlaneChart:           controlPlaneChart,
		ImageVectorFunc:             awsimagevector.GetImageVector,
		ConfigChartValuesFunc:       getConfigChartValues,
		ControlPlaneChartValuesFunc: getControlPlaneChartValues,
	}
}

var (
	controlPlaneSecrets = &secrets.Secrets{
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

	configChart = &chart.Chart{
		Name: "cloud-provider-config",
		Path: filepath.Join(aws.InternalChartsPath, "cloud-provider-config"),
		Objects: []*chart.Object{
			{Type: &corev1.ConfigMap{}, Name: common.CloudProviderConfigName},
		},
	}

	controlPlaneChart = &chart.Chart{
		Name:   "cloud-controller-manager",
		Path:   filepath.Join(aws.InternalChartsPath, "cloud-controller-manager"),
		Images: []string{common.HyperkubeImageName},
		Objects: []*chart.Object{
			{Type: &corev1.Service{}, Name: "cloud-controller-manager"},
			{Type: &appsv1.Deployment{}, Name: "cloud-controller-manager"},
		},
	}
)

// getConfigChartValues collects and returns the configuration chart values.
func getConfigChartValues(decoder runtime.Decoder, cp *extensionsv1alpha1.ControlPlane, cluster *extensionscontroller.Cluster) (map[string]interface{}, error) {
	// Decode providerConfig
	cpConfig := &apisaws.ControlPlaneConfig{}
	if _, _, err := decoder.Decode(cp.Spec.ProviderConfig.Raw, nil, cpConfig); err != nil {
		return nil, errors.Wrapf(err, "could not decode providerConfig of controlplane '%s'", util.ObjectName(cp))
	}

	// Decode infrastructureProviderStatus
	infraStatus := &apisaws.InfrastructureStatus{}
	if _, _, err := decoder.Decode(cp.Spec.InfrastructureProviderStatus.Raw, nil, infraStatus); err != nil {
		return nil, errors.Wrapf(err, "could not decode infrastructureProviderStatus of controlplane '%s'", util.ObjectName(cp))
	}

	// Determine subnet ID and zone
	subnetID, zone, err := getSubnetIDAndZone(infraStatus)
	if err != nil {
		return nil, errors.Wrapf(err, "could not determine subnet ID or zone from infrastructureProviderStatus of controlplane '%s'", util.ObjectName(cp))
	}

	// Collect config chart values
	return map[string]interface{}{
		"cloudProviderConfig": getCloudProviderConfig(infraStatus.VPC.ID, subnetID, zone, cp.Namespace),
	}, nil
}

// getControlPlaneChartValues collects and returns the control plane chart values.
func getControlPlaneChartValues(decoder runtime.Decoder, cp *extensionsv1alpha1.ControlPlane, cluster *extensionscontroller.Cluster, checksums map[string]string) (map[string]interface{}, error) {
	// Decode providerConfig
	cpConfig := &apisaws.ControlPlaneConfig{}
	if _, _, err := decoder.Decode(cp.Spec.ProviderConfig.Raw, nil, cpConfig); err != nil {
		return nil, errors.Wrapf(err, "could not decode providerConfig of controlplane '%s'", util.ObjectName(cp))
	}

	// Collect control plane chart values
	values := map[string]interface{}{
		"cloudProvider":     "aws",
		"clusterName":       cp.Namespace,
		"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
		"podNetwork":        extensionscontroller.GetPodNetwork(cluster.Shoot),
		"replicas":          extensionscontroller.GetReplicas(cluster.Shoot, 1),
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
	}
	if cpConfig.CloudControllerManager != nil {
		values["featureGates"] = cpConfig.CloudControllerManager.FeatureGates
	}
	return values, nil
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
