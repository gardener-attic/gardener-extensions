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
	"path/filepath"

	apisgcp "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/gcp"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/apihelper"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane/genericactuator"
	"github.com/gardener/gardener-extensions/pkg/util"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/authentication/user"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
					APIServerURL: gardencorev1alpha1.DeploymentNameKubeAPIServer,
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
	Path: filepath.Join(internal.InternalChartsPath, "cloud-provider-config"),
	Objects: []*chart.Object{
		{
			Type: &corev1.ConfigMap{},
			Name: internal.CloudProviderConfigName,
		},
	},
}

var ccmChart = &chart.Chart{
	Name:   "cloud-controller-manager",
	Path:   filepath.Join(internal.InternalChartsPath, "cloud-controller-manager"),
	Images: []string{gcp.HyperkubeImageName},
	Objects: []*chart.Object{
		{Type: &corev1.Service{}, Name: "cloud-controller-manager"},
		{Type: &appsv1.Deployment{}, Name: "cloud-controller-manager"},
		{Type: &corev1.ConfigMap{}, Name: "cloud-controller-manager-monitoring-config"},
	},
}

var ccmShootChart = &chart.Chart{
	Name: "cloud-controller-manager-shoot",
	Path: filepath.Join(internal.InternalChartsPath, "cloud-controller-manager-shoot"),
	Objects: []*chart.Object{
		{Type: &rbacv1.ClusterRole{}, Name: "system:controller:cloud-node-controller"},
		{Type: &rbacv1.ClusterRoleBinding{}, Name: "system:controller:cloud-node-controller"},
	},
}

var storageClassChart = &chart.Chart{
	Name: "shoot-storageclasses",
	Path: filepath.Join(gcp.InternalChartsPath, "shoot-storageclasses"),
}

// NewValuesProvider creates a new ValuesProvider for the generic actuator.
func NewValuesProvider(logger logr.Logger) genericactuator.ValuesProvider {
	return &valuesProvider{
		logger: logger.WithName("gcp-values-provider"),
	}
}

// valuesProvider is a ValuesProvider that provides AWS-specific values for the 2 charts applied by the generic actuator.
type valuesProvider struct {
	genericactuator.NoopValuesProvider
	decoder runtime.Decoder
	client  client.Client
	logger  logr.Logger
}

// InjectScheme injects the given scheme into the valuesProvider.
func (vp *valuesProvider) InjectScheme(scheme *runtime.Scheme) error {
	vp.decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
	return nil
}

// InjectClient injects the given client into the valuesProvider.
func (vp *valuesProvider) InjectClient(client client.Client) error {
	vp.client = client
	return nil
}

// GetConfigChartValues returns the values for the config chart applied by the generic actuator.
func (vp *valuesProvider) GetConfigChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) (map[string]interface{}, error) {
	// Decode providerConfig
	cpConfig := &apisgcp.ControlPlaneConfig{}
	if _, _, err := vp.decoder.Decode(cp.Spec.ProviderConfig.Raw, nil, cpConfig); err != nil {
		return nil, errors.Wrapf(err, "could not decode providerConfig of controlplane '%s'", util.ObjectName(cp))
	}

	// Decode infrastructureProviderStatus
	infraStatus := &apisgcp.InfrastructureStatus{}
	if _, _, err := vp.decoder.Decode(cp.Spec.InfrastructureProviderStatus.Raw, nil, infraStatus); err != nil {
		return nil, errors.Wrapf(err, "could not decode infrastructureProviderStatus of controlplane '%s'", util.ObjectName(cp))
	}

	// Get service account
	serviceAccount, err := internal.GetServiceAccount(ctx, vp.client, cp.Spec.SecretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get service account from secret '%s/%s'", cp.Spec.SecretRef.Namespace, cp.Spec.SecretRef.Name)
	}

	// Get config chart values
	return getConfigChartValues(cpConfig, infraStatus, cp, serviceAccount)
}

// GetControlPlaneChartValues returns the values for the control plane chart applied by the generic actuator.
func (vp *valuesProvider) GetControlPlaneChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksums map[string]string,
	scaledDown bool,
) (map[string]interface{}, error) {
	// Decode providerConfig
	cpConfig := &apisgcp.ControlPlaneConfig{}
	if _, _, err := vp.decoder.Decode(cp.Spec.ProviderConfig.Raw, nil, cpConfig); err != nil {
		return nil, errors.Wrapf(err, "could not decode providerConfig of controlplane '%s'", util.ObjectName(cp))
	}

	// Get CCM chart values
	return getCCMChartValues(cpConfig, cp, cluster, checksums, scaledDown)
}

// getConfigChartValues collects and returns the configuration chart values.
func getConfigChartValues(
	cpConfig *apisgcp.ControlPlaneConfig,
	infraStatus *apisgcp.InfrastructureStatus,
	cp *extensionsv1alpha1.ControlPlane,
	serviceAccount *internal.ServiceAccount,
) (map[string]interface{}, error) {
	// Determine network names
	networkName, subNetworkName := getNetworkNames(infraStatus, cp)

	// Collect config chart values
	return map[string]interface{}{
		"projectID":      serviceAccount.ProjectID,
		"networkName":    networkName,
		"subNetworkName": subNetworkName,
		"zone":           cpConfig.Zone,
		"nodeTags":       cp.Namespace,
	}, nil
}

// getCCMChartValues collects and returns the CCM chart values.
func getCCMChartValues(
	cpConfig *apisgcp.ControlPlaneConfig,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksums map[string]string,
	scaledDown bool,
) (map[string]interface{}, error) {
	values := map[string]interface{}{
		"replicas":          extensionscontroller.GetControlPlaneReplicas(cluster.Shoot, scaledDown, 1),
		"clusterName":       cp.Namespace,
		"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
		"podNetwork":        extensionscontroller.GetPodNetwork(cluster.Shoot),
		"podAnnotations": map[string]interface{}{
			"checksum/secret-cloud-controller-manager":        checksums[cloudControllerManagerDeploymentName],
			"checksum/secret-cloud-controller-manager-server": checksums[cloudControllerManagerServerName],
			"checksum/secret-cloudprovider":                   checksums[gardencorev1alpha1.SecretNameCloudProvider],
			"checksum/configmap-cloud-provider-config":        checksums[internal.CloudProviderConfigName],
		},
	}

	if cpConfig.CloudControllerManager != nil {
		values["featureGates"] = cpConfig.CloudControllerManager.FeatureGates
	}

	return values, nil
}

// getNetworkNames determines the network and sub-network names from the given infrastructure status and controlplane.
func getNetworkNames(
	infraStatus *apisgcp.InfrastructureStatus,
	cp *extensionsv1alpha1.ControlPlane,
) (string, string) {
	networkName := infraStatus.Networks.VPC.Name
	if networkName == "" {
		networkName = cp.Namespace
	}

	subNetworkName := ""
	subnet, _ := apihelper.FindSubnetForPurpose(infraStatus.Networks.Subnets, apisgcp.PurposeInternal)
	if subnet != nil {
		subNetworkName = subnet.Name
	}

	return networkName, subNetworkName
}
