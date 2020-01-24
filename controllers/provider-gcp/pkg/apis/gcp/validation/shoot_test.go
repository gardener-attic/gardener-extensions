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
	. "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp/validation"
	"github.com/gardener/gardener/pkg/apis/garden"
	"k8s.io/apimachinery/pkg/util/sets"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func makeStringPointer(s string) *string {
	ptr := s
	return &ptr
}

var _ = Describe("Shoot validation", func() {
	Describe("#ValidateNetworking", func() {
		var networkingPath = field.NewPath("spec", "networking")

		It("should return no error because nodes CIDR was provided", func() {
			networking := garden.Networking{
				Nodes: makeStringPointer("1.2.3.4/5"),
			}

			errorList := ValidateNetworking(networking, networkingPath)

			Expect(errorList).To(BeEmpty())
		})

		It("should return an error because no nodes CIDR was provided", func() {
			networking := garden.Networking{}

			errorList := ValidateNetworking(networking, networkingPath)

			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("spec.networking.nodes"),
				})),
			))
		})
	})

	Describe("#ValidateWorkers", func() {
		var (
			workers      []garden.Worker
			allowedZones sets.String
		)

		BeforeEach(func() {
			workers = []garden.Worker{
				{
					Volume: &garden.Volume{
						Type: makeStringPointer("Volume"),
						Size: "30G",
					},
					Zones: []string{
						"zone1",
						"zone2",
					},
				},
				{
					Volume: &garden.Volume{
						Type: makeStringPointer("Volume"),
						Size: "20G",
					},
					Zones: []string{
						"zone2",
						"zone3",
					},
				},
			}

			allowedZones = sets.NewString("zone1", "zone2", "zone3")

		})
		It("should pass because workers are configured correctly", func() {
			errorList := ValidateWorkers(workers, allowedZones, field.NewPath(""))

			Expect(errorList).To(BeEmpty())
		})

		It("should forbid because volume is not configured", func() {
			workers[1].Volume = nil

			errorList := ValidateWorkers(workers, allowedZones, field.NewPath("workers"))

			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("workers[1].volume"),
				})),
			))
		})

		It("should forbid because volume type and size are not configured", func() {
			workers[0].Volume.Type = nil
			workers[0].Volume.Size = ""

			errorList := ValidateWorkers(workers, allowedZones, field.NewPath("workers"))

			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("workers[0].volume.type"),
				})),
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("workers[0].volume.size"),
				})),
			))
		})

		It("should forbid because worker does not specify a zone", func() {
			workers[0].Zones = nil

			errorList := ValidateWorkers(workers, allowedZones, field.NewPath("workers"))

			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("workers[0].zones"),
				})),
			))
		})

		It("should forbid because worker use zones which are not available", func() {
			workers[0].Zones[0] = ""
			workers[1].Zones[1] = "not-available"

			errorList := ValidateWorkers(workers, allowedZones, field.NewPath("workers"))

			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("workers[0].zones[0]"),
				})),
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("workers[1].zones[1]"),
				})),
			))
		})
	})
})
