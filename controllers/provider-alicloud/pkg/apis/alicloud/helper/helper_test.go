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
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud"
	. "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helper", func() {
	var (
		purpose      alicloud.Purpose = "foo"
		purposeWrong alicloud.Purpose = "baz"
	)

	DescribeTable("#FindVSwitchForPurposeAndZone",
		func(vswitches []alicloud.VSwitch, purpose alicloud.Purpose, zone string, expectedVSwitch *alicloud.VSwitch, expectErr bool) {
			subnet, err := FindVSwitchForPurposeAndZone(vswitches, purpose, zone)
			expectResults(subnet, expectedVSwitch, err, expectErr)
		},

		Entry("list is nil", nil, purpose, "europe", nil, true),
		Entry("empty list", []alicloud.VSwitch{}, purpose, "europe", nil, true),
		Entry("entry not found (no purpose)", []alicloud.VSwitch{{ID: "bar", Purpose: purposeWrong, Zone: "europe"}}, purpose, "europe", nil, true),
		Entry("entry not found (no zone)", []alicloud.VSwitch{{ID: "bar", Purpose: purposeWrong, Zone: "europe"}}, purpose, "asia", nil, true),
		Entry("entry exists", []alicloud.VSwitch{{ID: "bar", Purpose: purposeWrong, Zone: "europe"}}, purposeWrong, "europe", &alicloud.VSwitch{ID: "bar", Purpose: purposeWrong, Zone: "europe"}, false),
	)

	DescribeTable("#FindSecurityGroupByPurpose",
		func(securityGroups []alicloud.SecurityGroup, purpose alicloud.Purpose, expectedSecurityGroup *alicloud.SecurityGroup, expectErr bool) {
			securityGroup, err := FindSecurityGroupByPurpose(securityGroups, purpose)
			expectResults(securityGroup, expectedSecurityGroup, err, expectErr)
		},

		Entry("list is nil", nil, purpose, nil, true),
		Entry("empty list", []alicloud.SecurityGroup{}, purpose, nil, true),
		Entry("entry not found", []alicloud.SecurityGroup{{ID: "bar", Purpose: purposeWrong}}, purpose, nil, true),
		Entry("entry exists", []alicloud.SecurityGroup{{ID: "bar", Purpose: purpose}}, purpose, &alicloud.SecurityGroup{ID: "bar", Purpose: purpose}, false),
	)

	DescribeTable("#FindMachineImage",
		func(machineImages []alicloud.MachineImage, name, version string, expectedMachineImage *alicloud.MachineImage, expectErr bool) {
			machineImage, err := FindMachineImage(machineImages, name, version)
			expectResults(machineImage, expectedMachineImage, err, expectErr)
		},

		Entry("list is nil", nil, "foo", "1.2.3", nil, true),
		Entry("empty list", []alicloud.MachineImage{}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no name)", []alicloud.MachineImage{{Name: "bar", Version: "1.2.3", ID: "id123"}}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no version)", []alicloud.MachineImage{{Name: "bar", Version: "1.2.3", ID: "id123"}}, "foo", "1.2.4", nil, true),
		Entry("entry exists", []alicloud.MachineImage{{Name: "bar", Version: "1.2.3", ID: "id123"}}, "bar", "1.2.3", &alicloud.MachineImage{Name: "bar", Version: "1.2.3", ID: "id123"}, false),
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
