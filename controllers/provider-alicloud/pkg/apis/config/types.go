// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package config

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfig "k8s.io/component-base/config"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration defines the configuration for the Alicloud provider.
type ControllerConfiguration struct {
	metav1.TypeMeta

	// ClientConnection specifies the kubeconfig file and client connection
	// settings for the proxy server to use when communicating with the apiserver.
	ClientConnection *componentbaseconfig.ClientConnectionConfiguration
	// MachineImages is the list of machine images that are understood by the controller. It maps
	// logical names and versions to Alicloud-specific identifiers.
	MachineImages []MachineImage
	// MachineImageOwnerSecretRef is the secret reference which contains credential of AliCloud subaccount for customized images.
	// We currently assume multiple customized images should always be under this account.
	MachineImageOwnerSecretRef *corev1.SecretReference
	// ETCD is the etcd configuration.
	ETCD ETCD
}

// MachineImage is a mapping from logical names and versions to Alicloud-specific identifiers.
type MachineImage struct {
	// Name is the logical name of the machine image.
	Name string
	// Version is the logical version of the machine image.
	Version string
	// Regions is a mapping to the correct IDs for the machine image in the supported regions.
	Regions []RegionImageMapping
}

// RegionImageMapping is a mapping from Region name to supported machine image id for a specific OS version.
type RegionImageMapping struct {
	// Region is the ID of the region.
	Region string
	// ImageID is the ID for the machine image.
	ImageID string
}

// ETCD is an etcd configuration.
type ETCD struct {
	// ETCDStorage is the etcd storage configuration.
	Storage ETCDStorage
	// ETCDBackup is the etcd backup configuration.
	Backup ETCDBackup
}

// ETCDStorage is an etcd storage configuration.
type ETCDStorage struct {
	// ClassName is the name of the storage class used in etcd-main volume claims.
	ClassName *string
	// Capacity is the storage capacity used in etcd-main volume claims.
	Capacity *resource.Quantity
}

// ETCDBackup is an etcd backup configuration.
type ETCDBackup struct {
	// Schedule is the etcd backup schedule.
	Schedule *string
}
