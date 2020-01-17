// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("ControlPlaneConfig validation", func() {
	var (
		region = "foo"
		zone   = "some-zone"

		regions = []gardencorev1beta1.Region{
			{
				Name: region,
				Zones: []gardencorev1beta1.AvailabilityZone{
					{Name: zone},
				},
			},
		}

		constraints = apisvsphere.Constraints{
			LoadBalancerConfig: apisvsphere.LoadBalancerConfig{
				Size: "MEDIUM",
				Classes: []apisvsphere.LoadBalancerClass{
					{
						Name:       "default",
						IPPoolName: "lbpool",
					},
					{
						Name:       "public",
						IPPoolName: "lbpool2",
					},
				},
			},
		}

		controlPlane *apisvsphere.ControlPlaneConfig
	)

	BeforeEach(func() {
		controlPlane = &apisvsphere.ControlPlaneConfig{
			LoadBalancerClasses: []apisvsphere.CPLoadBalancerClass{
				{Name: "default"},
			},
		}
	})

	Describe("#ValidateControlPlaneConfig", func() {
		It("should return no errors for a valid configuration", func() {
			Expect(ValidateControlPlaneConfig(controlPlane, region, regions, constraints)).To(BeEmpty())
		})

		It("should require the name of a load balancer class", func() {
			controlPlane.LoadBalancerClasses[0].Name = ""

			errorList := ValidateControlPlaneConfig(controlPlane, region, regions, constraints)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("loadBalancerClasses.name"),
			}))))
		})

		It("should require the name to be defined in the constraints", func() {
			controlPlane.LoadBalancerClasses[0].Name = "bar"

			errorList := ValidateControlPlaneConfig(controlPlane, region, regions, constraints)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("loadBalancerClasses.name"),
			}))))
		})
	})

	Describe("#ValidateControlPlaneConfigUpdate", func() {
		It("should return no errors for an unchanged config", func() {
			Expect(ValidateControlPlaneConfigUpdate(controlPlane, controlPlane, region, regions)).To(BeEmpty())
		})
	})
})
