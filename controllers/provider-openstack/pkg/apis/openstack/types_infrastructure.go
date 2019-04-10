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

import (
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta
	// FloatingPoolName contains the FloatingPoolName name in which LoadBalancer FIPs should be created.
	FloatingPoolName string
	// Networks is the OpenStack specific network configuration
	Networks Networks
	// Zones belonging to the same region
	Zones []Zone
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureStatus contains information about created infrastructure resources.
type InfrastructureStatus struct {
	metav1.TypeMeta

	// // Network contains information about the created Network and some related resources.
	// Network NetworkStatus

	// Router contains information about the Router and related resources.
	Router RouterStatus
}

// RouterStatus contains information about a generated Router or resources attached to an existing Router.
type RouterStatus struct {
	// ID is the Router id.
	ID string
	// Subnets is a list of subnets that have been created.
	Subnets []Subnet
}

// NetworkStatus contains information about a generated Network or resources created in an existing Network.
type NetworkStatus struct {
	// ID is the Network id.
	ID string
	// Subnets is a list of subnets that have been created.
	Subnets []Subnet
	// SecurityGroups is a list of security groups that have been created.
	SecurityGroups []SecurityGroup
}

// Networks holds information about the Kubernetes and infrastructure networks.
type Networks struct {
	// Router indicates whether to use an existing router or create a new one.
	// +optional
	Router *Router
	// Workers is a list of CIDRs of worker subnets (private) to create (used for the VMs).
	Workers []gardencorev1alpha1.CIDR
}

// Router indicates whether to use an existing router or create a new one.
type Router struct {
	// ID is the router id of an existing OpenStack router.
	ID string
}

// Zone describes the properties of a zone
type Zone struct {
	// Name is the name for this zone.
	Name string
	// Workers is the  workers  subnet range  to create (used for the VMs).
	Workers gardencorev1alpha1.CIDR
}

const (
	// PurposeNodes is a constant describing that the respective resource is used for nodes.
	PurposeNodes string = "nodes"
)

// Subnet is an OpenStack subnet related to a Network.
type Subnet struct {
	// Purpose is a logical description of the subnet.
	Purpose string
	// ID is the subnet id.
	ID string
	// Zone is the availability zone into which the subnet has been created.
	Zone string
}

// SecurityGroup is an OpenStack security group related to a Network.
type SecurityGroup struct {
	// Purpose is a logical description of the security group.
	Purpose string
	// ID is the subnet id.
	ID string
}
