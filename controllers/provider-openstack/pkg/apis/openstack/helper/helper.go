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

package helper

import (
	"fmt"

	"github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack"
)

// FindSubnetByPurpose takes a list of subnets and tries to find the first entry
// whose purpose matches with the given purpose. If no such entry is found then an error will be
// returned.
func FindSubnetByPurpose(subnets []openstack.Subnet, purpose openstack.Purpose) (*openstack.Subnet, error) {
	for _, subnet := range subnets {
		if subnet.Purpose == purpose {
			return &subnet, nil
		}
	}
	return nil, fmt.Errorf("cannot find subnet with purpose %q", purpose)
}

// FindSecurityGroupByPurpose takes a list of security groups and tries to find the first entry
// whose purpose matches with the given purpose. If no such entry is found then an error will be
// returned.
func FindSecurityGroupByPurpose(securityGroups []openstack.SecurityGroup, purpose openstack.Purpose) (*openstack.SecurityGroup, error) {
	for _, securityGroup := range securityGroups {
		if securityGroup.Purpose == purpose {
			return &securityGroup, nil
		}
	}
	return nil, fmt.Errorf("cannot find security group with purpose %q", purpose)
}

// FindMachineImage takes a list of machine images and tries to find the first entry
// whose name, version, and zone matches with the given name, version, and cloud profile. If no such
// entry is found then an error will be returned.
func FindMachineImage(machineImages []openstack.MachineImage, name, version, cloudProfile string) (*openstack.MachineImage, error) {
	for _, machineImage := range machineImages {
		if machineImage.Name == name && machineImage.Version == version && machineImage.CloudProfile == cloudProfile {
			return &machineImage, nil
		}
	}
	return nil, fmt.Errorf("no machine image with name %q, version %q, cloudProfile %q found", name, version, cloudProfile)
}
