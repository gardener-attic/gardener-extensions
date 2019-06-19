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
	"encoding/base64"
	"encoding/json"
	"path/filepath"

	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud"
	apisalicloud "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane/genericactuator"
	"github.com/gardener/gardener-extensions/pkg/util"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/authentication/user"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
					Name:         "cloud-controller-manager",
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
					Name:         "csi-attacher",
					CommonName:   "system:csi-attacher",
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
					Name:         "csi-provisioner",
					CommonName:   "system:csi-provisioner",
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
					Name:         "csi-snapshotter",
					CommonName:   "system:csi-snapshotter",
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[gardencorev1alpha1.SecretNameCACluster],
				},
				KubeConfigRequest: &secrets.KubeConfigRequest{
					ClusterName:  clusterName,
					APIServerURL: common.KubeAPIServerDeploymentName,
				},
			},
		}
	},
}

var configChart = &chart.Chart{
	Name: "cloud-provider-config",
	Path: filepath.Join(alicloud.InternalChartsPath, "cloud-provider-config"),
	Objects: []*chart.Object{
		{
			Type: &corev1.ConfigMap{},
			Name: alicloud.CloudProviderConfigName,
		},
	},
}

var controlPlaneChart = &chart.Chart{
	Name: "seed-controlplane",
	Path: filepath.Join(alicloud.InternalChartsPath, "seed-controlplane"),
	SubCharts: []*chart.Chart{
		{
			Name:   "alicloud-cloud-controller-manager",
			Images: []string{alicloud.CloudControllerManagerImageName},
			Objects: []*chart.Object{
				{Type: &corev1.Service{}, Name: "cloud-controller-manager"},
				{Type: &appsv1.Deployment{}, Name: "cloud-controller-manager"},
			},
		},
		{
			Name:   "csi-alicloud",
			Images: []string{alicloud.CSIAttacherImageName, alicloud.CSIProvisionerImageName, alicloud.CSISnapshotterImageName, alicloud.CSIPluginImageName},
			Objects: []*chart.Object{
				{Type: &appsv1.Deployment{}, Name: "csi-plugin-controller"},
			},
		},
	},
}

var controlPlaneShootChart = &chart.Chart{
	Name: "shoot-system-components",
	Path: filepath.Join(alicloud.InternalChartsPath, "shoot-system-components"),
	SubCharts: []*chart.Chart{
		{
			Name: "alicloud-cloud-controller-manager",
			Objects: []*chart.Object{
				{Type: &rbacv1.ClusterRole{}, Name: "system:controller:cloud-node-controller"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "system:controller:cloud-node-controller"},
			},
		},
		{
			Name:   "csi-alicloud",
			Images: []string{alicloud.CSINodeDriverRegistrarImageName, alicloud.CSIPluginImageName},
			Objects: []*chart.Object{
				{Type: &appsv1.DaemonSet{}, Name: "csi-disk-plugin-alicloud"},
				{Type: &corev1.Secret{}, Name: "csi-diskplugin-alicloud"},
				{Type: &corev1.ServiceAccount{}, Name: "csi-disk-plugin-alicloud"},
				{Type: &rbacv1.ClusterRole{}, Name: "garden.sapcloud.io:psp:kube-system:csi-disk-plugin-alicloud"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "garden.sapcloud.io:psp:csi-disk-plugin-alicloud"},
				{Type: &corev1.ServiceAccount{}, Name: "csi-attacher"},
				{Type: &rbacv1.ClusterRole{}, Name: "garden.sapcloud.io:kube-system:csi-attacher"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "garden.sapcloud.io:csi-attacher"},
				{Type: &rbacv1.Role{}, Name: "csi-attacher"},
				{Type: &rbacv1.RoleBinding{}, Name: "csi-attacher"},
				{Type: &corev1.ServiceAccount{}, Name: "csi-provisioner"},
				{Type: &rbacv1.ClusterRole{}, Name: "garden.sapcloud.io:kube-system:csi-provisioner"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "garden.sapcloud.io:csi-provisioner"},
				{Type: &corev1.ServiceAccount{}, Name: "csi-snapshotter"},
				{Type: &rbacv1.ClusterRole{}, Name: "garden.sapcloud.io:kube-system:csi-snapshotter"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "garden.sapcloud.io:csi-snapshotter"},
				{Type: &policyv1beta1.PodSecurityPolicy{}, Name: "gardener.kube-system.csi-disk-plugin-alicloud"},
			},
		},
	},
}

// NewValuesProvider creates a new ValuesProvider for the generic actuator.
func NewValuesProvider(logger logr.Logger) genericactuator.ValuesProvider {
	return &valuesProvider{
		logger: logger.WithName("alicloud-values-provider"),
	}
}

// valuesProvider is a ValuesProvider that provides AWS-specific values for the 2 charts applied by the generic actuator.
type valuesProvider struct {
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
	// Decode infrastructureProviderStatus
	infraStatus := &apisalicloud.InfrastructureStatus{}
	if _, _, err := vp.decoder.Decode(cp.Spec.InfrastructureProviderStatus.Raw, nil, infraStatus); err != nil {
		return nil, errors.Wrapf(err, "could not decode infrastructureProviderStatus of controlplane '%s'", util.ObjectName(cp))
	}

	// Get credentials from the referenced secret
	credentials, err := vp.getCredentials(ctx, cp)
	if err != nil {
		return nil, err
	}

	// Get config chart values
	return getConfigChartValues(infraStatus, cp, credentials)
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
	cpConfig := &apisalicloud.ControlPlaneConfig{}
	if _, _, err := vp.decoder.Decode(cp.Spec.ProviderConfig.Raw, nil, cpConfig); err != nil {
		return nil, errors.Wrapf(err, "could not decode providerConfig of controlplane '%s'", util.ObjectName(cp))
	}

	// Get control plane chart values
	return getControlPlaneChartValues(cpConfig, cp, cluster, checksums, scaledDown)
}

// GetControlPlaneShootChartValues returns the values for the control plane shoot chart applied by the generic actuator.
func (vp *valuesProvider) GetControlPlaneShootChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) (map[string]interface{}, error) {
	// Get credentials from the referenced secret
	credentials, err := vp.getCredentials(ctx, cp)
	if err != nil {
		return nil, err
	}

	// Get control plane shoot chart values
	return getControlPlaneShootChartValues(cluster, credentials)
}

// getCredentials determines the credentials from the secret referenced in the ControlPlane resource.
func (vp *valuesProvider) getCredentials(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
) (*alicloud.Credentials, error) {
	secret, err := extensionscontroller.GetSecretByReference(ctx, vp.client, &cp.Spec.SecretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get secret by reference for controlplane '%s'", util.ObjectName(cp))
	}
	credentials, err := alicloud.ReadSecretCredentials(secret)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read credentials from secret '%s'", util.ObjectName(secret))
	}
	return credentials, nil
}

// cloudConfig wraps the settings for the Alicloud provider.
// See https://github.com/kubernetes/cloud-provider-alibaba-cloud/blob/master/cloud-controller-manager/alicloud.go
type cloudConfig struct {
	Global struct {
		KubernetesClusterTag string
		UID                  string `json:"uid"`
		VpcID                string `json:"vpcid"`
		Region               string `json:"region"`
		ZoneID               string `json:"zoneid"`
		VswitchID            string `json:"vswitchid"`

		AccessKeyID     string `json:"accessKeyID"`
		AccessKeySecret string `json:"accessKeySecret"`
	}
}

// getConfigChartValues collects and returns the configuration chart values.
func getConfigChartValues(
	infraStatus *apisalicloud.InfrastructureStatus,
	cp *extensionsv1alpha1.ControlPlane,
	credentials *alicloud.Credentials,
) (map[string]interface{}, error) {
	// Determine vswitch ID and zone
	vswitchID, zone, err := getVswitchIDAndZone(infraStatus)
	if err != nil {
		return nil, errors.Wrapf(err, "could not determine vswitch ID or zone from infrastructureProviderStatus of controlplane '%s'", util.ObjectName(cp))
	}

	// Initialize cloud config
	cfg := &cloudConfig{}
	cfg.Global.KubernetesClusterTag = cp.Namespace
	cfg.Global.VpcID = infraStatus.VPC.ID
	cfg.Global.ZoneID = zone
	cfg.Global.VswitchID = vswitchID
	cfg.Global.AccessKeyID = base64.StdEncoding.EncodeToString([]byte(credentials.AccessKeyID))
	cfg.Global.AccessKeySecret = base64.StdEncoding.EncodeToString([]byte(credentials.AccessKeySecret))
	cfg.Global.Region = cp.Spec.Region

	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal cloud config to JSON for controlplane '%s'", util.ObjectName(cp))
	}

	// Collect config chart values
	return map[string]interface{}{
		"cloudConfig": string(cfgJSON),
	}, nil
}

// getControlPlaneChartValues collects and returns the control plane chart values.
func getControlPlaneChartValues(
	cpConfig *apisalicloud.ControlPlaneConfig,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksums map[string]string,
	scaledDown bool,
) (map[string]interface{}, error) {
	values := map[string]interface{}{
		"alicloud-cloud-controller-manager": map[string]interface{}{
			"replicas":          extensionscontroller.GetControlPlaneReplicas(cluster.Shoot, scaledDown, 1),
			"clusterName":       cp.Namespace,
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
			"podNetwork":        extensionscontroller.GetPodNetwork(cluster.Shoot),
			"podAnnotations": map[string]interface{}{
				"checksum/secret-cloud-controller-manager": checksums["cloud-controller-manager"],
				// TODO Use constant from github.com/gardener/gardener/pkg/apis/core/v1alpha1 when available
				// See https://github.com/gardener/gardener/pull/930
				"checksum/secret-cloudprovider":            checksums[common.CloudProviderSecretName],
				"checksum/configmap-cloud-provider-config": checksums[alicloud.CloudProviderConfigName],
			},
		},
		"csi-alicloud": map[string]interface{}{
			"replicas":          extensionscontroller.GetReplicas(cluster.Shoot, 1),
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
			"regionID":          cp.Spec.Region,
			"podAnnotations": map[string]interface{}{
				"checksum/secret-csi-attacher":    checksums["csi-attacher"],
				"checksum/secret-csi-provisioner": checksums["csi-provisioner"],
				"checksum/secret-csi-snapshotter": checksums["csi-snapshotter"],
				"checksum/secret-cloudprovider":   checksums[common.CloudProviderSecretName],
			},
		},
	}

	if cpConfig.CloudControllerManager != nil {
		values["alicloud-cloud-controller-manager"].(map[string]interface{})["featureGates"] = cpConfig.CloudControllerManager.FeatureGates
	}

	return values, nil
}

// getControlPlaneShootChartValues collects and returns the control plane shoot chart values.
func getControlPlaneShootChartValues(
	cluster *extensionscontroller.Cluster,
	credentials *alicloud.Credentials,
) (map[string]interface{}, error) {
	values := map[string]interface{}{
		"csi-alicloud": map[string]interface{}{
			"credential": map[string]interface{}{
				"accessKeyID":     base64.StdEncoding.EncodeToString([]byte(credentials.AccessKeyID)),
				"accessKeySecret": base64.StdEncoding.EncodeToString([]byte(credentials.AccessKeySecret)),
			},
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
		},
	}

	return values, nil
}

// getVswitchIDAndZone determines the vswitch ID and zone from the given infrastructure status by looking for the first
// subnet with purpose "nodes".
func getVswitchIDAndZone(infraStatus *apisalicloud.InfrastructureStatus) (string, string, error) {
	for _, vswitch := range infraStatus.VPC.VSwitches {
		if vswitch.Purpose == apisalicloud.PurposeNodes {
			return vswitch.ID, vswitch.Zone, nil
		}
	}
	return "", "", errors.Errorf("vswitch with purpose 'nodes' not found")
}
