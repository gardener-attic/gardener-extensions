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

package helper_test

import (
	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/apis/packet"
	. "github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/apis/packet/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helper", func() {
	DescribeTable("#FindMachineImage",
		func(machineImages []packet.MachineImage, name, version string, expectedMachineImage *packet.MachineImage, expectErr bool) {
			machineImage, err := FindMachineImage(machineImages, name, version)
			expectResults(machineImage, expectedMachineImage, err, expectErr)
		},

		Entry("list is nil", nil, "foo", "1.2.3", nil, true),
		Entry("empty list", []packet.MachineImage{}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no name)", []packet.MachineImage{{Name: "bar", Version: "1.2.3", ID: "1234"}}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no version)", []packet.MachineImage{{Name: "bar", Version: "1.2.3", ID: "1234"}}, "foo", "1.2.4", nil, true),
		Entry("entry exists", []packet.MachineImage{{Name: "bar", Version: "1.2.3", ID: "1234"}}, "bar", "1.2.3", &packet.MachineImage{Name: "bar", Version: "1.2.3", ID: "1234"}, false),
	)
})

func expectResults(result, expected interface{}, err error, expectErr bool) {
	if !expectErr {
		Expect(result).To(Equal(expected))
		Expect(err).NotTo(HaveOccurred())
	} else {
		Expect(result).To(BeNil())
		Expect(err).To(HaveOccurred())
	}
}
