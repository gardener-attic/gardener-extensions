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

package validation

import (
	"fmt"
	"net"
	"time"

	apisopenstack "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateProviderProfileConfig validates a ProviderProfileConfig object.
func ValidateProviderProfileConfig(providerProfile *apisopenstack.ProviderProfileConfig) field.ErrorList {
	allErrs := field.ErrorList{}

	floatingPoolPath := field.NewPath("constraints", "floatingPools")
	if len(providerProfile.Constraints.FloatingPools) == 0 {
		allErrs = append(allErrs, field.Required(floatingPoolPath, "must provide at least one floating pool"))
	}
	for i, pool := range providerProfile.Constraints.FloatingPools {
		if len(pool.Name) == 0 {
			allErrs = append(allErrs, field.Required(floatingPoolPath.Index(i).Child("name"), "must provide a name"))
		}
	}

	loadBalancerProviderPath := field.NewPath("constraints", "loadBalancerProviders")
	if len(providerProfile.Constraints.LoadBalancerProviders) == 0 {
		allErrs = append(allErrs, field.Required(loadBalancerProviderPath, "must provide at least one load balancer provider"))
	}
	for i, pool := range providerProfile.Constraints.LoadBalancerProviders {
		if len(pool.Name) == 0 {
			allErrs = append(allErrs, field.Required(loadBalancerProviderPath.Index(i).Child("name"), "must provide a name"))
		}
	}

	machineImagesPath := field.NewPath("machineImages")
	if len(providerProfile.MachineImages) == 0 {
		allErrs = append(allErrs, field.Required(machineImagesPath, "must provide at least one machine image"))
	}
	for i, machineImage := range providerProfile.MachineImages {
		idxPath := machineImagesPath.Index(i)

		if len(machineImage.Name) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("name"), "must provide a name"))
		}

		if len(machineImage.Versions) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("versions"), fmt.Sprintf("must provide at least one version for machine image %q", machineImage.Name)))
		}
		for j, version := range machineImage.Versions {
			jdxPath := idxPath.Child("versions").Index(j)

			if len(version.Version) == 0 {
				allErrs = append(allErrs, field.Required(jdxPath.Child("version"), "must provide a version"))
			}
			if len(version.Image) == 0 {
				allErrs = append(allErrs, field.Required(jdxPath.Child("image"), "must provide an image"))
			}
		}
	}

	if len(providerProfile.KeyStoneURL) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("keyStoneURL"), "must provide the URL to KeyStone"))
	}

	for i, ip := range providerProfile.DNSServers {
		if net.ParseIP(ip) == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("dnsServers").Index(i), ip, "must provide a valid IP"))
		}
	}

	if providerProfile.DHCPDomain != nil && len(*providerProfile.DHCPDomain) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("dhcpDomain"), "must provide a dhcp domain when the key is specified"))
	}

	if providerProfile.RequestTimeout != nil {
		if _, err := time.ParseDuration(*providerProfile.RequestTimeout); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("requestTimeout"), *providerProfile.RequestTimeout, fmt.Sprintf("invalid duration: %v", err)))
		}
	}

	return allErrs
}
