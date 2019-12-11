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
	api "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud"
	. "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helper", func() {
	var (
		purpose      api.Purpose = "foo"
		purposeWrong api.Purpose = "baz"
	)

	DescribeTable("#FindVSwitchForPurposeAndZone",
		func(vswitches []api.VSwitch, purpose api.Purpose, zone string, expectedVSwitch *api.VSwitch, expectErr bool) {
			subnet, err := FindVSwitchForPurposeAndZone(vswitches, purpose, zone)
			expectResults(subnet, expectedVSwitch, err, expectErr)
		},

		Entry("list is nil", nil, purpose, "europe", nil, true),
		Entry("empty list", []api.VSwitch{}, purpose, "europe", nil, true),
		Entry("entry not found (no purpose)", []api.VSwitch{{ID: "bar", Purpose: purposeWrong, Zone: "europe"}}, purpose, "europe", nil, true),
		Entry("entry not found (no zone)", []api.VSwitch{{ID: "bar", Purpose: purposeWrong, Zone: "europe"}}, purpose, "asia", nil, true),
		Entry("entry exists", []api.VSwitch{{ID: "bar", Purpose: purposeWrong, Zone: "europe"}}, purposeWrong, "europe", &api.VSwitch{ID: "bar", Purpose: purposeWrong, Zone: "europe"}, false),
	)

	DescribeTable("#FindSecurityGroupByPurpose",
		func(securityGroups []api.SecurityGroup, purpose api.Purpose, expectedSecurityGroup *api.SecurityGroup, expectErr bool) {
			securityGroup, err := FindSecurityGroupByPurpose(securityGroups, purpose)
			expectResults(securityGroup, expectedSecurityGroup, err, expectErr)
		},

		Entry("list is nil", nil, purpose, nil, true),
		Entry("empty list", []api.SecurityGroup{}, purpose, nil, true),
		Entry("entry not found", []api.SecurityGroup{{ID: "bar", Purpose: purposeWrong}}, purpose, nil, true),
		Entry("entry exists", []api.SecurityGroup{{ID: "bar", Purpose: purpose}}, purpose, &api.SecurityGroup{ID: "bar", Purpose: purpose}, false),
	)

	DescribeTable("#FindMachineImage",
		func(configImages []api.MachineImage, name, version string, expectedMachineImage *api.MachineImage, expectErr bool) {
			machineImage, err := FindMachineImage(configImages, name, version)
			expectResults(machineImage, expectedMachineImage, err, expectErr)
		},

		Entry("list is nil", nil, "foo", "1.2.3", nil, true),
		Entry("empty list", []api.MachineImage{}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no name)", []api.MachineImage{{Name: "bar", Version: "1.2.3", ID: "id123"}}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no version)", []api.MachineImage{{Name: "bar", Version: "1.2.3", ID: "id123"}}, "foo", "1.2.4", nil, true),
		Entry("entry exists", []api.MachineImage{{Name: "bar", Version: "1.2.3", ID: "id123"}}, "bar", "1.2.3", &api.MachineImage{Name: "bar", Version: "1.2.3", ID: "id123"}, false),
	)
})

const profileImageID = "id-1235"

var _ = Describe("Helper", func() {
	DescribeTable("#FindImageForRegion",
		func(profileImages []api.MachineImages, imageName, version string, expectedImage string) {
			cfg := &api.CloudProfileConfig{}
			cfg.MachineImages = profileImages
			image, err := FindImageForRegionFromCloudProfile(cfg, imageName, version)

			Expect(image).To(Equal(expectedImage))
			if expectedImage != "" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},

		Entry("list is nil", nil, "ubuntu", "1", ""),

		Entry("profile empty list", []api.MachineImages{}, "ubuntu", "1", ""),
		Entry("profile entry not found (image does not exist)", makeProfileMachineImages("debian", "1"), "ubuntu", "1", ""),
		Entry("profile entry not found (version does not exist)", makeProfileMachineImages("ubuntu", "2"), "ubuntu", "1", ""),
		Entry("profile entry", makeProfileMachineImages("ubuntu", "1"), "ubuntu", "1", profileImageID),
	)
})

func makeProfileMachineImages(name, version string) []api.MachineImages {
	versions := []api.MachineImageVersion{
		api.MachineImageVersion{
			Version: version,
			ID:      profileImageID,
		},
	}

	return []api.MachineImages{
		{
			Name:     name,
			Versions: versions,
		},
	}
}

func expectResults(result, expected interface{}, err error, expectErr bool) {
	if !expectErr {
		Expect(result).To(Equal(expected))
		Expect(err).NotTo(HaveOccurred())
	} else {
		Expect(result).To(BeNil())
		Expect(err).To(HaveOccurred())
	}
}
