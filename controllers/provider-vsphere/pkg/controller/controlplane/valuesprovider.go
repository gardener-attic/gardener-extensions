/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package controlplane

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	apisvsphere "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere"
	apishelper "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere/helper"
	"github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/internal"
	"github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/internal/helper"
	"github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/vsphere"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/common"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane/genericactuator"

	v1alpha1constants "github.com/gardener/gardener/pkg/apis/core/v1alpha1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apiserver/pkg/authentication/user"
)

// Object names
const (
	cloudControllerManagerServerName = "cloud-controller-manager-server"
)

var controlPlaneSecrets = &secrets.Secrets{
	CertificateSecretConfigs: map[string]*secrets.CertificateSecretConfig{
		v1alpha1constants.SecretNameCACluster: {
			Name:       v1alpha1constants.SecretNameCACluster,
			CommonName: "kubernetes",
			CertType:   secrets.CACert,
		},
	},
	SecretConfigsFunc: func(cas map[string]*secrets.Certificate, clusterName string) []secrets.ConfigInterface {
		return []secrets.ConfigInterface{
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         vsphere.CloudControllerManagerName,
					CommonName:   "system:serviceaccount:kube-system:cloud-controller-manager",
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1alpha1constants.SecretNameCACluster],
				},
				KubeConfigRequest: &secrets.KubeConfigRequest{
					ClusterName:  clusterName,
					APIServerURL: v1alpha1constants.DeploymentNameKubeAPIServer,
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:       cloudControllerManagerServerName,
					CommonName: vsphere.CloudControllerManagerName,
					DNSNames:   controlplane.DNSNamesForService(vsphere.CloudControllerManagerName, clusterName),
					CertType:   secrets.ServerCert,
					SigningCA:  cas[v1alpha1constants.SecretNameCACluster],
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         "csi-attacher",
					CommonName:   "system:csi-attacher",
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1alpha1constants.SecretNameCACluster],
				},
				KubeConfigRequest: &secrets.KubeConfigRequest{
					ClusterName:  clusterName,
					APIServerURL: v1alpha1constants.DeploymentNameKubeAPIServer,
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         "csi-provisioner",
					CommonName:   "system:csi-provisioner",
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1alpha1constants.SecretNameCACluster],
				},
				KubeConfigRequest: &secrets.KubeConfigRequest{
					ClusterName:  clusterName,
					APIServerURL: v1alpha1constants.DeploymentNameKubeAPIServer,
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         "vsphere-csi-controller",
					CommonName:   "gardener.cloud:vsphere-csi-controller",
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1alpha1constants.SecretNameCACluster],
				},
				KubeConfigRequest: &secrets.KubeConfigRequest{
					ClusterName:  clusterName,
					APIServerURL: v1alpha1constants.DeploymentNameKubeAPIServer,
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         "vsphere-csi-syncer",
					CommonName:   "gardener.cloud:vsphere-csi-syncer",
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1alpha1constants.SecretNameCACluster],
				},
				KubeConfigRequest: &secrets.KubeConfigRequest{
					ClusterName:  clusterName,
					APIServerURL: v1alpha1constants.DeploymentNameKubeAPIServer,
				},
			},
		}
	},
}

var configChart = &chart.Chart{
	Name: "cloud-provider-config",
	Path: filepath.Join(vsphere.InternalChartsPath, "cloud-provider-config"),
	Objects: []*chart.Object{
		{Type: &corev1.ConfigMap{}, Name: vsphere.CloudProviderConfig},
	},
}

var controlPlaneChart = &chart.Chart{
	Name: "seed-controlplane",
	Path: filepath.Join(vsphere.InternalChartsPath, "seed-controlplane"),
	SubCharts: []*chart.Chart{
		{
			Name:   "vsphere-cloud-controller-manager",
			Images: []string{vsphere.CloudControllerImageName},
			Objects: []*chart.Object{
				{Type: &corev1.Service{}, Name: "cloud-controller-manager"},
				{Type: &appsv1.Deployment{}, Name: "cloud-controller-manager"},
				{Type: &corev1.ConfigMap{}, Name: "cloud-controller-manager-monitoring-config"},
			},
		},
		{
			Name:   "nsxt-lb-provider-manager",
			Images: []string{vsphere.NsxtLbProviderImageName},
			Objects: []*chart.Object{
				{Type: &corev1.Service{}, Name: "lb-controller-manager"},
				{Type: &appsv1.Deployment{}, Name: "lb-controller-manager"},
			},
		},
		{
			Name: "csi-vsphere",
			Images: []string{vsphere.CSIAttacherImageName, vsphere.CSIProvisionerImageName, vsphere.CSIControllerImageName,
				vsphere.CSISyncerImageName, vsphere.LivenessProbeImageName},
			Objects: []*chart.Object{
				{Type: &corev1.Secret{}, Name: "csi-vsphere-config-secret"},
				{Type: &appsv1.StatefulSet{}, Name: "vsphere-csi-controller"},
			},
		},
	},
}

var controlPlaneShootChart = &chart.Chart{
	Name: "shoot-system-components",
	Path: filepath.Join(vsphere.InternalChartsPath, "shoot-system-components"),
	SubCharts: []*chart.Chart{
		{
			Name: "vsphere-cloud-controller-manager",
			Objects: []*chart.Object{
				{Type: &corev1.ServiceAccount{}, Name: "cloud-controller-manager"},
				{Type: &rbacv1.ClusterRole{}, Name: "system:cloud-controller-manager"},
				{Type: &rbacv1.RoleBinding{}, Name: "system:cloud-controller-manager:apiserver-authentication-reader"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "system:cloud-controller-manager"},
			},
		},
		{
			Name:   "csi-vsphere",
			Images: []string{vsphere.CSINodeDriverRegistrarImageName, vsphere.CSINodeImageName, vsphere.LivenessProbeImageName},
			Objects: []*chart.Object{
				{Type: &corev1.Secret{}, Name: "csi-vsphere-config-secret"},
				{Type: &rbacv1.ClusterRole{}, Name: "gardener.cloud:vsphere-csi-controller"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "gardener.cloud:vsphere-csi-controller"},
				{Type: &rbacv1.ClusterRole{}, Name: "gardener.cloud:vsphere-csi-syncer"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "gardener.cloud:vsphere-csi-syncer"},
				{Type: &rbacv1.ClusterRole{}, Name: "gardener.cloud:csi-attacher"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "gardener.cloud:csi-attacher"},
				{Type: &rbacv1.ClusterRole{}, Name: "gardener.cloud:csi-provisioner"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "gardener.cloud:csi-provisioner"},
				{Type: &appsv1.DaemonSet{}, Name: "vsphere-csi-node"},
			},
		},
	},
}

var storageClassChart = &chart.Chart{
	Name: "shoot-storageclasses",
	Path: filepath.Join(vsphere.InternalChartsPath, "shoot-storageclasses"),
}

// NewValuesProvider creates a new ValuesProvider for the generic actuator.
func NewValuesProvider(logger logr.Logger, gardenId string) genericactuator.ValuesProvider {
	return &valuesProvider{
		logger:   logger.WithName("vsphere-values-provider"),
		gardenId: gardenId,
	}
}

// valuesProvider is a ValuesProvider that provides vSphere-specific values for the 2 charts applied by the generic actuator.
type valuesProvider struct {
	genericactuator.NoopValuesProvider
	common.ClientContext
	logger   logr.Logger
	gardenId string
}

// GetConfigChartValues returns the values for the config chart applied by the generic actuator.
func (vp *valuesProvider) GetConfigChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) (map[string]interface{}, error) {
	cpConfig, err := helper.GetControlPlaneConfig(&vp.ClientContext, cluster)
	if err != nil {
		return nil, err
	}

	// Get credentials
	credentials, err := internal.GetCredentials(ctx, vp.Client(), cp.Spec.SecretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get vSphere credentials from secret '%s/%s'", cp.Spec.SecretRef.Namespace, cp.Spec.SecretRef.Name)
	}

	// Get config chart values
	return vp.getConfigChartValues(cp, cpConfig, cluster, credentials)
}

// GetControlPlaneChartValues returns the values for the control plane chart applied by the generic actuator.
func (vp *valuesProvider) GetControlPlaneChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksums map[string]string,
	scaledDown bool,
) (map[string]interface{}, error) {
	cpConfig, err := helper.GetControlPlaneConfig(&vp.ClientContext, cluster)
	if err != nil {
		return nil, err
	}

	// Get credentials
	credentials, err := internal.GetCredentials(ctx, vp.Client(), cp.Spec.SecretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get vSphere credentials from secret '%s/%s'", cp.Spec.SecretRef.Namespace, cp.Spec.SecretRef.Name)
	}

	// Get control plane chart values
	return vp.getControlPlaneChartValues(cpConfig, cp, cluster, credentials, checksums, scaledDown)
}

// GetControlPlaneShootChartValues returns the values for the control plane shoot chart applied by the generic actuator.
func (vp *valuesProvider) GetControlPlaneShootChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksums map[string]string,
) (map[string]interface{}, error) {
	// Get credentials
	credentials, err := internal.GetCredentials(ctx, vp.Client(), cp.Spec.SecretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get vSphere credentials from secret '%s/%s'", cp.Spec.SecretRef.Namespace, cp.Spec.SecretRef.Name)
	}

	// Get control plane shoot chart values
	return vp.getControlPlaneShootChartValues(cp, cluster, credentials)
}

// GetStorageClassesChartValues returns the values for the shoot storageclasses chart applied by the generic actuator.
func (vp *valuesProvider) GetStorageClassesChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) (map[string]interface{}, error) {

	cloudProfileConfig, err := helper.GetCloudProfileConfig(&vp.ClientContext, cluster)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"storagePolicyName": cloudProfileConfig.DefaultClassStoragePolicyName,
	}, nil
}

func splitServerNameAndPort(host string) (name string, port int, err error) {
	parts := strings.Split(host, ":")
	if len(parts) == 1 {
		name = host
		port = 443
	} else if len(parts) == 2 {
		name = parts[0]
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, errors.Wrapf(err, "invalid port for vSphere host: host=%s,port=%s", host, parts[1])
		}
	} else {
		return "", 0, fmt.Errorf("invalid vSphere host: %s (too many parts %v)", host, parts)
	}

	return
}

// getConfigChartValues collects and returns the configuration chart values.
func (vp *valuesProvider) getConfigChartValues(
	cp *extensionsv1alpha1.ControlPlane,
	cpConfig *apisvsphere.ControlPlaneConfig,
	cluster *extensionscontroller.Cluster,
	credentials *internal.Credentials,
) (map[string]interface{}, error) {

	cloudProfileConfig, err := helper.GetCloudProfileConfig(&vp.ClientContext, cluster)
	if err != nil {
		return nil, err
	}

	region := apishelper.FindRegion(cluster.Shoot.Spec.Region, cloudProfileConfig)
	if region == nil {
		return nil, fmt.Errorf("region %q not found in cloud profile config", cluster.Shoot.Spec.Region)
	}

	insecureFlag := "0"
	if region.VsphereInsecureSSL {
		insecureFlag = "1"
	}

	serverName, port, err := splitServerNameAndPort(region.VsphereHost)
	if err != nil {
		return nil, err
	}

	infraStatus, err := helper.GetInfrastructureStatus(&vp.ClientContext, cp.Name, cp.Spec.InfrastructureProviderStatus)
	if err != nil {
		return nil, err
	}

	var defaultClass *apisvsphere.LoadBalancerClass
	loadBalancersClasses := []map[string]interface{}{}
	if len(cpConfig.LoadBalancerClasses) == 0 {
		cpConfig.LoadBalancerClasses = []apisvsphere.CPLoadBalancerClass{{Name: apisvsphere.LoadBalancerDefaultClassName}}
	}
	for i, class := range cloudProfileConfig.Constraints.LoadBalancerConfig.Classes {
		if i == 0 || class.Name == apisvsphere.LoadBalancerDefaultClassName {
			class0 := class
			defaultClass = &class0
		}
	}
outer:
	for _, cpClass := range cpConfig.LoadBalancerClasses {
		lbClass := map[string]interface{}{
			"name": cpClass.Name,
		}
		if cpClass.IPPoolName == nil || *cpClass.IPPoolName == "" {
			for _, class := range cloudProfileConfig.Constraints.LoadBalancerConfig.Classes {
				if class.Name == cpClass.Name {
					if class.IPPoolName != "" {
						lbClass["ipPoolName"] = class.IPPoolName
					}
					loadBalancersClasses = append(loadBalancersClasses, lbClass)
					continue outer
				}
			}
			return nil, fmt.Errorf("load balancer class %q not found in cloud profile", cpClass.Name)
		} else {
			lbClass["ipPoolName"] = *cpClass.IPPoolName
			loadBalancersClasses = append(loadBalancersClasses, lbClass)
		}
	}

	if defaultClass.IPPoolName == "" {
		return nil, fmt.Errorf("load balancer default class %q must specify both ipPoolName and size in cloud profile", defaultClass.Name)
	}

	loadBalancers := map[string]interface{}{
		"ipPoolName": defaultClass.IPPoolName,
		"size":       cloudProfileConfig.Constraints.LoadBalancerConfig.Size,
		"classes":    loadBalancersClasses,
	}
	if infraStatus.LogicalRouterId != "" {
		loadBalancers["logicalRouterId"] = infraStatus.LogicalRouterId
	}

	// Collect config chart values
	values := map[string]interface{}{
		"serverName":   serverName,
		"serverPort":   port,
		"insecureFlag": insecureFlag,
		"datacenters":  strings.Join(apishelper.CollectDatacenters(region), ","),
		"username":     credentials.VsphereUsername,
		"password":     credentials.VspherePassword,
		"loadbalancer": loadBalancers,
		"nsxt": map[string]interface{}{
			"host":         region.NSXTHost,
			"insecureFlag": region.VsphereInsecureSSL,
			"username":     credentials.NSXTUsername,
			"password":     credentials.NSXTPassword,
		},
	}

	if region.CaFile != nil && *region.CaFile != "" {
		values["caFile"] = *region.CaFile
	}
	if region.Thumbprint != nil && *region.Thumbprint != "" {
		values["thumbprint"] = *region.Thumbprint
	}
	if cloudProfileConfig.FailureDomainLabels != nil {
		values["labelRegion"] = cloudProfileConfig.FailureDomainLabels.Region
		values["labelZone"] = cloudProfileConfig.FailureDomainLabels.Zone
	}

	return values, nil
}

// getControlPlaneChartValues collects and returns the control plane chart values.
func (vp *valuesProvider) getControlPlaneChartValues(
	cpConfig *apisvsphere.ControlPlaneConfig,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	credentials *internal.Credentials,
	checksums map[string]string,
	scaledDown bool,
) (map[string]interface{}, error) {

	cloudProfileConfig, err := helper.GetCloudProfileConfig(&vp.ClientContext, cluster)
	if err != nil {
		return nil, err
	}

	region := apishelper.FindRegion(cluster.Shoot.Spec.Region, cloudProfileConfig)
	if region == nil {
		return nil, fmt.Errorf("region %q not found in cloud profile config", cluster.Shoot.Spec.Region)
	}

	insecureFlag := "false"
	if region.VsphereInsecureSSL {
		insecureFlag = "true"
	}

	serverName, port, err := splitServerNameAndPort(region.VsphereHost)
	if err != nil {
		return nil, err
	}

	clusterId := cp.Namespace + "-" + vp.gardenId
	values := map[string]interface{}{
		"vsphere-cloud-controller-manager": map[string]interface{}{
			"replicas":          extensionscontroller.GetControlPlaneReplicas(cluster, scaledDown, 1),
			"clusterName":       clusterId,
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
			"podNetwork":        extensionscontroller.GetPodNetwork(cluster),
			"podAnnotations": map[string]interface{}{
				"checksum/secret-cloud-controller-manager":        checksums[vsphere.CloudControllerManagerName],
				"checksum/secret-cloud-controller-manager-server": checksums[cloudControllerManagerServerName],
				"checksum/secret-cloudprovider":                   checksums[v1alpha1constants.SecretNameCloudProvider],
				"checksum/configmap-cloud-provider-config":        checksums[vsphere.CloudProviderConfig],
			},
		},
		"nsxt-lb-provider-manager": map[string]interface{}{
			"replicas":          extensionscontroller.GetControlPlaneReplicas(cluster, scaledDown, 1),
			"clusterName":       clusterId,
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
			"podAnnotations": map[string]interface{}{
				"checksum/secret-cloud-controller-manager":        checksums[vsphere.CloudControllerManagerName],
				"checksum/secret-cloud-controller-manager-server": checksums[cloudControllerManagerServerName],
				"checksum/secret-cloudprovider":                   checksums[v1alpha1constants.SecretNameCloudProvider],
				"checksum/configmap-cloud-provider-config":        checksums[vsphere.CloudProviderConfig],
			},
		},
		"csi-vsphere": map[string]interface{}{
			"replicas":          extensionscontroller.GetControlPlaneReplicas(cluster, scaledDown, 1),
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
			"serverName":        serverName,
			"clusterID":         clusterId,
			"username":          credentials.VsphereUsername,
			"password":          credentials.VspherePassword,
			"serverPort":        port,
			"datacenters":       strings.Join(apishelper.CollectDatacenters(region), ","),
			"insecureFlag":      insecureFlag,
			"podAnnotations": map[string]interface{}{
				"checksum/secret-csi-attacher":              checksums["csi-attacher"],
				"checksum/secret-csi-provisioner":           checksums["csi-provisioner"],
				"checksum/secret-vsphere-csi-controller":    checksums["vsphere-csi-controller"],
				"checksum/secret-vsphere-csi-syncer":        checksums["csi-vsphere-csi-syncer"],
				"checksum/secret-cloudprovider":             checksums[v1alpha1constants.SecretNameCloudProvider],
				"checksum/secret-csi-vsphere-config-secret": checksums[vsphere.SecretCsiVsphereConfig],
			},
		},
	}

	if cpConfig.CloudControllerManager != nil {
		values["vsphere-cloud-controller-manager"].(map[string]interface{})["featureGates"] = cpConfig.CloudControllerManager.FeatureGates
	}

	return values, nil
}

// getControlPlaneShootChartValues collects and returns the control plane shoot chart values.
func (vp *valuesProvider) getControlPlaneShootChartValues(
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	credentials *internal.Credentials,
) (map[string]interface{}, error) {

	cloudProfileConfig, err := helper.GetCloudProfileConfig(&vp.ClientContext, cluster)
	if err != nil {
		return nil, err
	}

	region := apishelper.FindRegion(cluster.Shoot.Spec.Region, cloudProfileConfig)
	if region == nil {
		return nil, fmt.Errorf("region %q not found in cloud profile config", cluster.Shoot.Spec.Region)
	}

	insecureFlag := "false"
	if region.VsphereInsecureSSL {
		insecureFlag = "true"
	}

	serverName, port, err := splitServerNameAndPort(region.VsphereHost)
	if err != nil {
		return nil, err
	}

	clusterId := cp.Namespace + "-" + vp.gardenId
	values := map[string]interface{}{
		"csi-vsphere": map[string]interface{}{
			"serverName":        serverName,
			"clusterID":         clusterId,
			"username":          credentials.VsphereUsername,
			"password":          credentials.VspherePassword,
			"serverPort":        port,
			"datacenters":       strings.Join(apishelper.CollectDatacenters(region), ","),
			"insecureFlag":      insecureFlag,
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
		},
	}

	return values, nil
}
