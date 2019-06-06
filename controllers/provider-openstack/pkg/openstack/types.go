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

package openstack

import "path/filepath"

const (
	// Name is the name of the OpenStack provider.
	Name = "provider-openstack"

	// MachineControllerManagerImageName is the name of the MachineControllerManager image.
	MachineControllerManagerImageName = "machine-controller-manager"
	// HyperkubeImageName is the name of the hyperkube image.
	HyperkubeImageName = "hyperkube"

	// AuthURL is a constant for the key in a cloud provider secret that holds the OpenStack auth url.
	AuthURL = "authURL"

	// CloudProviderConfigName is the name of the configmap containing the cloud provider config.
	CloudProviderConfigName = "cloud-provider-config"

	// DomainName is a constant for the key in a cloud provider secret that holds the OpenStack domain name.
	DomainName = "domainName"
	// TenantName is a constant for the key in a cloud provider secret that holds the OpenStack tenant name.
	TenantName = "tenantName"
	// UserName is a constant for the key in a cloud provider secret and backup secret that holds the OpenStack username.
	UserName = "username"
	// Password is a constant for the key in a cloud provider secret and backup secret that holds the OpenStack password.
	Password = "password"

	// MachineControllerManagerName is a constant for the name of the machine-controller-manager.
	MachineControllerManagerName = "machine-controller-manager"
)

var (
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("controllers", Name, "charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")
)
