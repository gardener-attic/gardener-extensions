// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://wwr.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package worker_test

import (
	"github.com/gardener/gardener-extensions/pkg/controller/worker"

	"k8s.io/apimachinery/pkg/util/intstr"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Machines", func() {
	Context("MachineDeployment", func() {
		DescribeTable("#HasDeployment",
			func(machineDeployments worker.MachineDeployments, name string, expectation bool) {
				Expect(machineDeployments.HasDeployment(name)).To(Equal(expectation))
			},

			Entry("list is nil", nil, "foo", false),
			Entry("empty list", worker.MachineDeployments{}, "foo", false),
			Entry("entry not found", worker.MachineDeployments{{Name: "bar"}}, "foo", false),
			Entry("entry exists", worker.MachineDeployments{{Name: "bar"}}, "bar", true),
		)

		DescribeTable("#HasClass",
			func(machineDeployments worker.MachineDeployments, class string, expectation bool) {
				Expect(machineDeployments.HasClass(class)).To(Equal(expectation))
			},

			Entry("list is nil", nil, "foo", false),
			Entry("empty list", worker.MachineDeployments{}, "foo", false),
			Entry("entry not found", worker.MachineDeployments{{ClassName: "bar"}}, "foo", false),
			Entry("entry exists", worker.MachineDeployments{{ClassName: "bar"}}, "bar", true),
		)

		DescribeTable("#HasSecret",
			func(machineDeployments worker.MachineDeployments, secret string, expectation bool) {
				Expect(machineDeployments.HasSecret(secret)).To(Equal(expectation))
			},

			Entry("list is nil", nil, "foo", false),
			Entry("empty list", worker.MachineDeployments{}, "foo", false),
			Entry("entry not found", worker.MachineDeployments{{SecretName: "bar"}}, "foo", false),
			Entry("entry exists", worker.MachineDeployments{{SecretName: "bar"}}, "bar", true),
		)
	})

	DescribeTable("#MachineClassHash",
		func(spec map[string]interface{}, version, expectedHash string) {
			Expect(worker.MachineClassHash(spec, version)).To(Equal(expectedHash))
		},

		Entry("empty spec", nil, "", "aa969"),
		Entry("non-empty spec", map[string]interface{}{"foo": "bar"}, "1.5", "5a88b"),
	)

	DescribeTable("#DistributeOverZones",
		func(zoneIndex, size, zoneSize, expectation int) {
			Expect(worker.DistributeOverZones(zoneIndex, size, zoneSize)).To(Equal(expectation))
		},

		Entry("one zone, size 5", 0, 5, 1, 5),
		Entry("two zones, size 5, first index", 0, 5, 2, 3),
		Entry("two zones, size 5, second index", 1, 5, 2, 2),
		Entry("two zones, size 6, first index", 0, 6, 2, 3),
		Entry("two zones, size 6, second index", 1, 6, 2, 3),
		Entry("three zones, size 9, first index", 0, 9, 3, 3),
		Entry("three zones, size 9, second index", 1, 9, 3, 3),
		Entry("three zones, size 9, third index", 2, 9, 3, 3),
		Entry("three zones, size 10, first index", 0, 10, 3, 4),
		Entry("three zones, size 10, second index", 1, 10, 3, 3),
		Entry("three zones, size 10, third index", 2, 10, 3, 3),
	)

	DescribeTable("#DistributePercentOverZones",
		func(zoneIndex int, percent string, zoneSize, total int, expectation string) {
			Expect(worker.DistributePercentOverZones(zoneIndex, percent, zoneSize, total)).To(Equal(expectation))
		},

		Entry("even size, size 2", 0, "10%", 2, 8, "10%"),
		Entry("even size, size 2", 1, "50%", 2, 2, "50%"),
		Entry("uneven size, size 2", 0, "50%", 2, 5, "60%"),
		Entry("uneven size, size 2", 1, "50%", 2, 5, "40%"),
		Entry("uneven size, size 3", 0, "75%", 3, 5, "90%"),
		Entry("uneven size, size 3", 1, "75%", 3, 5, "90%"),
		Entry("uneven size, size 3", 2, "75%", 3, 5, "45%"),
	)

	DescribeTable("#DistributePositiveIntOrPercent",
		func(zoneIndex int, intOrPercent intstr.IntOrString, zoneSize, total int, expectation intstr.IntOrString) {
			Expect(worker.DistributePositiveIntOrPercent(zoneIndex, intOrPercent, zoneSize, total)).To(Equal(expectation))
		},

		Entry("percent", 2, intstr.FromString("75%"), 3, 5, intstr.FromString("45%")),
		Entry("positive int", 2, intstr.FromInt(10), 3, 3, intstr.FromInt(3)),
	)

	DescribeTable("#DiskSize",
		func(size string, expectation int, errMatcher types.GomegaMatcher) {
			val, err := worker.DiskSize(size)

			Expect(val).To(Equal(expectation))
			Expect(err).To(errMatcher)
		},

		Entry("1-digit size", "2Gi", 2, BeNil()),
		Entry("2-digit size", "20Gi", 20, BeNil()),
		Entry("3-digit size", "200Gi", 200, BeNil()),
		Entry("4-digit size", "2000Gi", 2000, BeNil()),
		Entry("non-parseable size", "foo", -1, HaveOccurred()),
	)
})
