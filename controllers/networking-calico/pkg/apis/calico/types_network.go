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

package calico

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Backend string

const (
	Bird Backend = "bird"
	None Backend = "none"
)

type CIDR string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkConfig configuration for the calico networking plugin
type NetworkConfig struct {
	metav1.TypeMeta
	// Backend defines whether a backend should be used or not (e.g., bird or None)
	Backend Backend
	// IPAM to use for the Calico Plugin (e.g., host-local or Calico)
	// +optional
	IPAM *IPAM
	// IPAutoDetectionMethod is the method to use to autodetect the IPv4 address for this host. This is only used when the IPv4 address is being autodetected.
	// https://docs.projectcalico.org/v2.2/reference/node/configuration#ip-autodetection-methods
	// +optional
	IPAutoDetectionMethod *string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkStatus contains information about created Network resources.
type NetworkStatus struct {
	metav1.TypeMeta
}

// IPAM defines the block that configuration for the ip assignment plugin to be used
type IPAM struct {
	// Type defines the IPAM plugin type
	Type string
	// CIDR defines the CIDR block to be used
	CIDR *CIDR
}
