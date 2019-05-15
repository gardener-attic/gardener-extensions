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

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"
)

// FindInstanceProfileForPurpose takes a list of instance profiles and tries to find the first entry
// whose purpose matches with the given purpose. If no such entry is found then an error will be
// returned.
func FindInstanceProfileForPurpose(instanceProfiles []aws.InstanceProfile, purpose string) (*aws.InstanceProfile, error) {
	for _, instanceProfile := range instanceProfiles {
		if instanceProfile.Purpose == purpose {
			return &instanceProfile, nil
		}
	}
	return nil, fmt.Errorf("no instance profile with purpose %q found", purpose)
}

// FindRoleForPurpose takes a list of roles and tries to find the first entry
// whose purpose matches with the given purpose. If no such entry is found then an error will be
// returned.
func FindRoleForPurpose(roles []aws.Role, purpose string) (*aws.Role, error) {
	for _, role := range roles {
		if role.Purpose == purpose {
			return &role, nil
		}
	}
	return nil, fmt.Errorf("no role with purpose %q found", purpose)
}

// FindSecurityGroupForPurpose takes a list of security groups and tries to find the first entry
// whose purpose matches with the given purpose. If no such entry is found then an error will be
// returned.
func FindSecurityGroupForPurpose(securityGroups []aws.SecurityGroup, purpose string) (*aws.SecurityGroup, error) {
	for _, securityGroup := range securityGroups {
		if securityGroup.Purpose == purpose {
			return &securityGroup, nil
		}
	}
	return nil, fmt.Errorf("no security group with purpose %q found", purpose)
}

// FindSubnetForPurposeAndZone takes a list of subnets and tries to find the first entry
// whose purpose and zone matches with the given purpose and zone. If no such entry is found then
// an error will be returned.
func FindSubnetForPurposeAndZone(subnets []aws.Subnet, purpose, zone string) (*aws.Subnet, error) {
	for _, subnet := range subnets {
		if subnet.Purpose == purpose && subnet.Zone == zone {
			return &subnet, nil
		}
	}
	return nil, fmt.Errorf("no subnet with purpose %q in zone %q found", purpose, zone)
}
