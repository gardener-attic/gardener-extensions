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
	apisaws "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"
	. "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws/validation"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("ValidateWorkerConfig", func() {
	var (
		workers  []gardencorev1alpha1.Worker
		awsZones []apisaws.Zone
	)

	BeforeEach(func() {
		workers = []gardencorev1alpha1.Worker{
			gardencorev1alpha1.Worker{
				Volume: &gardencorev1alpha1.Volume{
					Type: makeStringPointer("Volume"),
					Size: "30G",
				},
				Zones: []string{
					"zone1",
					"zone2",
				},
			},
			gardencorev1alpha1.Worker{
				Volume: &gardencorev1alpha1.Volume{
					Type: makeStringPointer("Volume"),
					Size: "20G",
				},
				Zones: []string{
					"zone2",
					"zone3",
				},
			},
		}

		awsZones = []apisaws.Zone{
			apisaws.Zone{
				Name: "zone1",
			},
			apisaws.Zone{
				Name: "zone2",
			},
			apisaws.Zone{
				Name: "zone3",
			},
		}
	})

	Describe("#ValidateWorkers", func() {
		It("should pass because workers are configured correctly", func() {
			errorList := ValidateWorkers(workers, awsZones, field.NewPath(""))

			Expect(errorList).To(BeEmpty())
		})

		It("should forbid because volume is not configured", func() {
			workers[1].Volume = nil

			errorList := ValidateWorkers(workers, awsZones, field.NewPath("workers"))

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

			errorList := ValidateWorkers(workers, awsZones, field.NewPath("workers"))

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

			errorList := ValidateWorkers(workers, awsZones, field.NewPath("workers"))

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

			errorList := ValidateWorkers(workers, awsZones, field.NewPath("workers"))

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

func makeStringPointer(s string) *string {
	ptr := s
	return &ptr
}
