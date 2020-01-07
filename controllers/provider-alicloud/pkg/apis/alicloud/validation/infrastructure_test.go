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

package validation_test

import (
	apisalicloud "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud"
	. "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/validation"

	. "github.com/gardener/gardener/pkg/utils/validation/gomega"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("InfrastructureConfig validation", func() {
	var (
		infrastructureConfig *apisalicloud.InfrastructureConfig

		pods        = "100.96.0.0/11"
		services    = "100.64.0.0/13"
		nodes       = "10.250.0.0/16"
		vpc         = "10.0.0.0/8"
		invalidCIDR = "invalid-cidr"
	)

	BeforeEach(func() {
		infrastructureConfig = &apisalicloud.InfrastructureConfig{
			Networks: apisalicloud.Networks{
				VPC: apisalicloud.VPC{
					CIDR: &vpc,
				},
				Zones: []apisalicloud.Zone{
					{
						Name:    "zone1",
						Workers: "10.250.3.0/24",
					},
				},
			},
		}
	})

	Describe("#ValidateInfrastructureConfig", func() {
		Context("CIDR", func() {
			It("should forbid invalid VPC CIDRs", func() {
				infrastructureConfig.Networks.VPC.CIDR = &invalidCIDR

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.vpc.cidr"),
					"Detail": Equal("invalid CIDR address: invalid-cidr"),
				}))
			})

			It("should forbid invalid workers CIDR", func() {
				infrastructureConfig.Networks.Zones[0].Workers = invalidCIDR

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.zones[0].workers"),
					"Detail": Equal("invalid CIDR address: invalid-cidr"),
				}))
			})

			It("should forbid workers CIDR which are not in Nodes CIDR", func() {
				infrastructureConfig.Networks.Zones[0].Workers = "1.1.1.1/32"

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.zones[0].workers"),
					"Detail": Equal(`must be a subset of "" ("10.250.0.0/16")`),
				}, Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.zones[0].workers"),
					"Detail": Equal(`must be a subset of "networks.vpc.cidr" ("10.0.0.0/8")`),
				}))
			})

			It("should forbid Node which are not in VPC CIDR", func() {
				notOverlappingCIDR := "1.1.1.1/32"
				infrastructureConfig.Networks.Zones[0].Workers = notOverlappingCIDR

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &notOverlappingCIDR, &pods, &services)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal(""),
					"Detail": Equal(`must be a subset of "networks.vpc.cidr" ("10.0.0.0/8")`),
				}, Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.zones[0].workers"),
					"Detail": Equal(`must be a subset of "networks.vpc.cidr" ("10.0.0.0/8")`),
				}))
			})

			It("should forbid Pod CIDR to overlap with VPC CIDR", func() {
				podCIDR := "10.0.0.1/32"

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &podCIDR, &services)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal(""),
					"Detail": Equal(`must not be a subset of "networks.vpc.cidr" ("10.0.0.0/8")`),
				}))
			})

			It("should forbid Services CIDR to overlap with VPC CIDR", func() {
				servicesCIDR := "10.0.0.1/32"

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &servicesCIDR)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal(""),
					"Detail": Equal(`must not be a subset of "networks.vpc.cidr" ("10.0.0.0/8")`),
				}))
			})

			It("should forbid non canonical CIDRs", func() {
				var (
					vpcCIDR     = "10.0.0.3/8"
					nodeCIDR    = "10.250.0.3/16"
					podCIDR     = "100.96.0.4/11"
					serviceCIDR = "100.64.0.5/13"
				)

				infrastructureConfig.Networks.Zones[0].Workers = "10.250.3.8/24"
				infrastructureConfig.Networks.VPC = apisalicloud.VPC{CIDR: &vpcCIDR}

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodeCIDR, &podCIDR, &serviceCIDR)

				Expect(errorList).To(HaveLen(2))
				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.vpc.cidr"),
					"Detail": Equal("must be valid canonical CIDR"),
				}, Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.zones[0].workers"),
					"Detail": Equal("must be valid canonical CIDR"),
				}))
			})
		})
	})

	Describe("#ValidateInfrastructureConfigUpdate", func() {
		It("should return no errors for an unchanged config", func() {
			Expect(ValidateInfrastructureConfigUpdate(infrastructureConfig, infrastructureConfig, &nodes, &pods, &services)).To(BeEmpty())
		})

		It("should forbid changing the network section", func() {
			newInfrastructureConfig := infrastructureConfig.DeepCopy()
			newCIDR := "1.2.3.4/5"
			newInfrastructureConfig.Networks.VPC.CIDR = &newCIDR

			errorList := ValidateInfrastructureConfigUpdate(infrastructureConfig, newInfrastructureConfig, &nodes, &pods, &services)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("networks"),
			}))))
		})
	})
})
