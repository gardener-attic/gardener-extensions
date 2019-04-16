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
	// TerrformerPurposeInfra is a constant for the complete Terraform setup with purpose 'infrastructure'.
	TerrformerPurposeInfra = "infra"
	// SSHKeyName key for accessing SSH key name from outputs in terraform
	SSHKeyName = "key_name"
	// RouterID is the id the router between provider network and the worker subnet.
	RouterID = "router_id"
	// NetworkID is the private worker network.
	NetworkID = "network_id"
	// SecurityGroupID is the id of worker security group.
	SecurityGroupID = "security_group_id"
	// SecurityGroupName is the name of the worker security group.
	SecurityGroupName = "security_group_name"
	// FloatingNetworkID is the id of the provider network.
	FloatingNetworkID = "floating_network_id"
	// SubnetID is the id of the worker subnet.
	SubnetID = "subnet_id"
)

var (
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("controllers", "provider-openstack", "charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")
)
