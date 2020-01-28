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

package vsphere

import "path/filepath"

const (
	// Name is the name of the vSphere provider controller.
	Name = "provider-vsphere"

	// TerraformerImageName is the name of the Terraformer image.
	TerraformerImageName = "terraformer"
	// MachineControllerManagerImageName is the name of the MachineControllerManager image.
	MachineControllerManagerImageName = "machine-controller-manager"
	// MCMProviderVsphereImageName is the namne of the vSphere provider plugin image.
	MCMProviderVsphereImageName = "machine-controller-manager-provider-vsphere"
	// CloudControllerImageName is the name of the external vSphere CloudProvider image.
	CloudControllerImageName = "vsphere-cloud-controller-manager"
	// NsxtLbProviderImageName is the name of the external NSX-T Load Balancer Provider image.
	NsxtLbProviderImageName = "nsxt-lb-provider-manager"

	// CSIAttacherImageName is the name of the CSI attacher image.
	CSIAttacherImageName = "csi-attacher"
	// CSINodeDriverRegistrarImageName is the name of the CSI driver registrar image.
	CSINodeDriverRegistrarImageName = "csi-node-driver-registrar"
	// CSIProvisionerImageName is the name of the CSI provisioner image.
	CSIProvisionerImageName = "csi-provisioner"
	// CSIControllerImageName is the name of the CSI plugin image.
	CSIControllerImageName = "vsphere-csi-controller"
	// CSISyncerImageName is the name of the vSphere CSI Syncer image.
	CSISyncerImageName = "vsphere-csi-syncer"
	// LivenessProbeImageName is the name of the liveness-probe image.
	LivenessProbeImageName = "liveness-probe"
	// CSINodeImageName is the name of the vsphere-csi-node image.
	CSINodeImageName = "vsphere-csi-node"

	// Host is a constant for the key in a cloud provider secret holding the VSphere host name
	Host = "vsphereHost"
	// Username is a constant for the key in a cloud provider secret holding the VSphere user name
	Username = "vsphereUsername"
	// Password is a constant for the key in a cloud provider secret holding the VSphere password
	Password = "vspherePassword"
	// InsecureSSL is a constant for the key in a cloud provider secret holding the boolean flag to allow insecure HTTPS connections to the VSphere host
	InsecureSSL = "vsphereInsecureSSL"

	// NSXTUsername is a constant for the key in a cloud provider secret holding the NSX-T user name
	NSXTUsername = "nsxtUsername"
	// Password is a constant for the key in a cloud provider secret holding the NSX-T password
	NSXTPassword = "nsxtPassword"

	// TerraformerPurposeInfra is a constant for the complete Terraform setup with purpose 'infrastructure'.
	TerraformerPurposeInfra = "infra"

	// CloudProviderConfig is the name of the configmap containing the cloud provider config.
	CloudProviderConfig = "cloud-provider-config"
	// CloudProviderConfigMapKey is the key storing the cloud provider config as value in the cloud provider configmap.
	CloudProviderConfigMapKey = "cloudprovider.conf"
	// SecretCsiVsphereConfig is a constant for the secret containing the CSI vSphere config.
	SecretCsiVsphereConfig = "csi-vsphere-config-secret"
	// MachineControllerManagerName is a constant for the name of the machine-controller-manager.
	MachineControllerManagerName = "machine-controller-manager"
	// MachineControllerManagerVpaName is the name of the VerticalPodAutoscaler of the machine-controller-manager deployment.
	MachineControllerManagerVpaName = "machine-controller-manager-vpa"
	// MachineControllerManagerMonitoringConfigName is the name of the ConfigMap containing monitoring stack configurations for machine-controller-manager.
	MachineControllerManagerMonitoringConfigName = "machine-controller-manager-monitoring-config"

	// CloudControllerManagerName is the constant for the name of the CloudController deployed by the control plane controller.
	CloudControllerManagerName = "cloud-controller-manager"
)

var (
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("controllers", Name, "charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")
)
