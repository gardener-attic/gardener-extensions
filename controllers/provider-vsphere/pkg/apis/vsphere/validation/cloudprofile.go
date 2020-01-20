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

package validation

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/sets"
	"strings"

	apisvsphere "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

var validValues = sets.NewString("SMALL", "MEDIUM", "LARGE")

// ValidateCloudProfileConfig validates a CloudProfileConfig object.
func ValidateCloudProfileConfig(cloudProfile *apisvsphere.CloudProfileConfig) field.ErrorList {
	allErrs := field.ErrorList{}

	loadBalancerSizePath := field.NewPath("constraints", "loadBalancerConfig", "size")
	if cloudProfile.Constraints.LoadBalancerConfig.Size == "" {
		allErrs = append(allErrs, field.Required(loadBalancerSizePath, "must provide the load balancer size"))
	} else {
		if !validValues.Has(cloudProfile.Constraints.LoadBalancerConfig.Size) {
			allErrs = append(allErrs, field.Required(loadBalancerSizePath,
				fmt.Sprintf("must provide a valid load balancer size value (%s)", strings.Join(validValues.List(), ","))))
		}
	}

	if cloudProfile.NamePrefix == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("namePrefix"), "must provide name prefix for NSX-T resources"))
	}
	if cloudProfile.DefaultClassStoragePolicyName == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("defaultClassStoragePolicyName"), "must provide defaultClassStoragePolicyName"))
	}

	machineImagesPath := field.NewPath("machineImages")
	if len(cloudProfile.MachineImages) == 0 {
		allErrs = append(allErrs, field.Required(machineImagesPath, "must provide at least one machine image"))
	}

	checkMachineImage := func(idxPath *field.Path, machineImage apisvsphere.MachineImages) {
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
			if len(version.Path) == 0 {
				allErrs = append(allErrs, field.Required(jdxPath.Child("path"), "must provide a path of VM template"))
			}
		}
	}
	for i, machineImage := range cloudProfile.MachineImages {
		checkMachineImage(machineImagesPath.Index(i), machineImage)
	}

	regionsPath := field.NewPath("regions")
	if len(cloudProfile.Regions) == 0 {
		allErrs = append(allErrs, field.Required(regionsPath, "must provide at least one region"))
	}
	for i, region := range cloudProfile.Regions {
		regionPath := regionsPath.Index(i)
		if region.Name == "" {
			allErrs = append(allErrs, field.Required(regionPath.Child("name"), "must provide region name"))
		}
		if region.VsphereHost == "" {
			allErrs = append(allErrs, field.Required(regionPath.Child("vsphereHost"), fmt.Sprintf("must provide vSphere host for region %s", region.Name)))
		}
		if region.NSXTHost == "" {
			allErrs = append(allErrs, field.Required(regionPath.Child("nsxtHost"), fmt.Sprintf("must provide NSX-T  host for region %s", region.Name)))
		}
		if region.SNATIPPool == "" {
			allErrs = append(allErrs, field.Required(regionPath.Child("snatIPPool"), fmt.Sprintf("must provide SNAT IP pool for region %s", region.Name)))
		}
		if region.TransportZone == "" {
			allErrs = append(allErrs, field.Required(regionPath.Child("transportZone"), fmt.Sprintf("must provide transport zone for region %s", region.Name)))
		}
		if region.LogicalTier0Router == "" {
			allErrs = append(allErrs, field.Required(regionPath.Child("logicalTier0Router"), fmt.Sprintf("must provide logical tier 0 router for region %s", region.Name)))
		}
		if region.EdgeCluster == "" {
			allErrs = append(allErrs, field.Required(regionPath.Child("edgeCluster"), fmt.Sprintf("must provide edge cluster for region %s", region.Name)))
		}
		if len(region.Zones) == 0 {
			allErrs = append(allErrs, field.Required(regionPath.Child("zones"), fmt.Sprintf("must provide edge cluster for region %s", region.Name)))
		}
		if len(cloudProfile.DNSServers) == 0 && len(region.DNSServers) == 0 {
			allErrs = append(allErrs, field.Required(field.NewPath("dnsServers"), "must provide dnsServers globally or for each region"))
			allErrs = append(allErrs, field.Required(regionPath.Child("dnsServers"), fmt.Sprintf("must provide dnsServers globally or for region %s", region.Name)))
		}
		for j, zone := range region.Zones {
			zonePath := regionPath.Child("zones").Index(j)
			if zone.Name == "" {
				allErrs = append(allErrs, field.Required(zonePath.Child("name"), fmt.Sprintf("must provide zone name in zones for region %s", region.Name)))
			}
			if !isSet(zone.Datacenter) && !isSet(region.Datacenter) {
				allErrs = append(allErrs, field.Required(zonePath.Child("datacenter"), fmt.Sprintf("must provide data center either for region %s or its zone %s", region.Name, zone.Name)))
			}
			if !isSet(zone.Datastore) && !isSet(zone.DatastoreCluster) && !isSet(region.Datastore) && !isSet(region.DatastoreCluster) {
				allErrs = append(allErrs, field.Required(zonePath.Child("datastore"), fmt.Sprintf("must provide either data store or data store cluster for either region %s or its zone %s", region.Name, zone.Name)))
			}
			if !isSet(zone.ComputeCluster) && !isSet(zone.ResourcePool) && !isSet(zone.HostSystem) {
				allErrs = append(allErrs, field.Required(zonePath.Child("resourcePool"), fmt.Sprintf("must provide either compute cluster, resource pool, or hostsystem for region %s, zone %s", region.Name, zone.Name)))
			}
		}
		for i, machineImage := range region.MachineImages {
			checkMachineImage(regionPath.Child("machineImages").Index(i), machineImage)
		}
	}

	return allErrs
}

func isSet(s *string) bool {
	return s != nil && *s != ""
}
