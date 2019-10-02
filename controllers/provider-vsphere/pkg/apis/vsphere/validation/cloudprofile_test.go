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

package validation_test

import (
	apisvsphere "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere"
	. "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere/validation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("ValidateCloudProfileConfig", func() {
	Describe("#ValidateCloudProfileConfig", func() {
		var cloudProfileConfig *apisvsphere.CloudProfileConfig

		BeforeEach(func() {
			cloudProfileConfig = &apisvsphere.CloudProfileConfig{
				NamePrefix:                    "prefix",
				DefaultClassStoragePolicyName: "default-class",
				Constraints: apisvsphere.Constraints{
					LoadBalancerConfig: apisvsphere.LoadBalancerConfig{
						Size: "MEDIUM",
						Classes: []apisvsphere.LoadBalancerClass{
							{
								Name:       "default",
								IPPoolName: "lbpool",
							},
						},
					},
				},
				DNSServers: []string{"1.2.3.4"},
				Regions: []apisvsphere.RegionSpec{
					{
						Name:               "region1",
						VsphereHost:        "vsphere.somewhere",
						NSXTHost:           "nsxt.somewhere",
						TransportZone:      "tz",
						EdgeCluster:        "edgecluster",
						LogicalTier0Router: "tier0",
						SNATIPPool:         "snat-pool",
						Zones: []apisvsphere.ZoneSpec{
							{
								Name:         "rz1",
								Datacenter:   "dc",
								Datastore:    "ds",
								ResourcePool: "mypool",
							},
						},
					},
				},
				MachineImages: []apisvsphere.MachineImages{
					{
						Name: "coreos",
						Versions: []apisvsphere.MachineImageVersion{
							{
								Version: "2190.5.0",
								Path:    "gardener/templates/coreos-2190.5.0",
								GuestID: "coreos64Guest",
							},
							{
								Version: "2190.5.1",
								Path:    "gardener/templates/coreos-2190.5.1",
							},
						},
					},
				},
			}
		})

		Context("machine image validation", func() {
			It("should validate valid machine image version configuration", func() {
				errorList := ValidateCloudProfileConfig(cloudProfileConfig)
				Expect(errorList).To(ConsistOf())
			})

			It("should validate valid machine image version configuration", func() {
				errorList := ValidateCloudProfileConfig(cloudProfileConfig)
				Expect(errorList).To(ConsistOf())
			})

			It("should enforce that at least one machine image has been defined", func() {
				cloudProfileConfig.MachineImages = []apisvsphere.MachineImages{}

				errorList := ValidateCloudProfileConfig(cloudProfileConfig)

				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("machineImages"),
				}))))
			})

			It("should forbid unsupported machine image configuration", func() {
				cloudProfileConfig.MachineImages = []apisvsphere.MachineImages{{}}

				errorList := ValidateCloudProfileConfig(cloudProfileConfig)

				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("machineImages[0].name"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("machineImages[0].versions"),
				}))))
			})

			It("should forbid unsupported machine image version configuration", func() {
				cloudProfileConfig.MachineImages = []apisvsphere.MachineImages{
					{
						Name:     "abc",
						Versions: []apisvsphere.MachineImageVersion{{}},
					},
				}

				errorList := ValidateCloudProfileConfig(cloudProfileConfig)

				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("machineImages[0].versions[0].version"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("machineImages[0].versions[0].path"),
				}))))
			})

			It("should have a load balancer size", func() {
				cloudProfileConfig.Constraints.LoadBalancerConfig.Size = ""

				errorList := ValidateCloudProfileConfig(cloudProfileConfig)

				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":   Equal(field.ErrorTypeRequired),
					"Field":  Equal("constraints.loadBalancerConfig.size"),
					"Detail": Equal("must provide the load balancer size"),
				}))))
			})

			It("should have a valid load balancer size value", func() {
				cloudProfileConfig.Constraints.LoadBalancerConfig.Size = "foo"

				errorList := ValidateCloudProfileConfig(cloudProfileConfig)

				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("constraints.loadBalancerConfig.size"),
				}))))
			})

			It("should have a valid compute cluster/resource pool/host system", func() {
				cloudProfileConfig.Regions[0].Zones[0].ResourcePool = ""

				cloudProfileConfig.Regions[0].Zones[0].ComputeCluster = "cc"
				errorList := ValidateCloudProfileConfig(cloudProfileConfig)
				Expect(errorList).To(ConsistOf())

				cloudProfileConfig.Regions[0].Zones[0].ComputeCluster = ""
				errorList = ValidateCloudProfileConfig(cloudProfileConfig)
				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("regions[0].zones[0].resourcePool"),
				}))))
			})

			It("should have a valid datastore", func() {
				cloudProfileConfig.Regions[0].Zones[0].Datastore = ""

				errorList := ValidateCloudProfileConfig(cloudProfileConfig)

				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("regions[0].zones[0].datastore"),
				}))))
			})

			It("should have a valid datacenter", func() {
				cloudProfileConfig.Regions[0].Zones[0].Datacenter = ""
				cloudProfileConfig.Regions[0].Datacenter = "dc"

				errorList := ValidateCloudProfileConfig(cloudProfileConfig)
				Expect(errorList).To(ConsistOf())

				cloudProfileConfig.Regions[0].Datacenter = ""
				errorList = ValidateCloudProfileConfig(cloudProfileConfig)
				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("regions[0].zones[0].datacenter"),
				}))))
			})
		})
	})
})
