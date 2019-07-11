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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControlPlaneConfig contains configuration settings for the control plane.
type ControlPlaneConfig struct {
	metav1.TypeMeta

	// LoadBalancerProvider is the name of the load balancer provider in the OpenStack environment.
	LoadBalancerProvider string

	// FloatingPoolClasses available for a dedicated Shoot.
	// +optional
	LoadBalancerClasses []LoadBalancerClass

	// CloudControllerManager contains configuration settings for the cloud-controller-manager.
	// +optional
	CloudControllerManager *CloudControllerManagerConfig
}

const (
	// DefaultLoadBalancerClass defines the default load balancer class.
	DefaultLoadBalancerClass = "default"
	// PrivateLoadBalancerClass defines the load balancer class used to default the private load balancers.
	PrivateLoadBalancerClass = "private"
	// VPNLoadBalancerClass defines the floating pool class used by the VPN service.
	VPNLoadBalancerClass = "vpn"
)

// LoadBalancerClass defines a restricted network setting for generic LoadBalancer classes usable in CloudProfiles.
type LoadBalancerClass struct {
	// Name is the name of the LB class
	Name string
	// FloatingSubnetID is the subnetwork ID of a dedicated subnet in floating network pool.
	// +optional
	FloatingSubnetID *string
	// FloatingNetworkID is the network ID of the floating network pool.
	// +optional
	FloatingNetworkID *string
	// SubnetID is the ID of a local subnet used for LoadBalancer provisioning. Only usable if no FloatingPool
	// configuration is done.
	// +optional
	SubnetID *string
}

// CloudControllerManagerConfig contains configuration settings for the cloud-controller-manager.
type CloudControllerManagerConfig struct {
	// FeatureGates contains information about enabled feature gates.
	// +optional
	FeatureGates map[string]bool `json:"featureGates,omitempty"`
}
