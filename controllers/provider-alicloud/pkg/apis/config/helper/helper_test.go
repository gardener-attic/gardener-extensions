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
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/config"
	. "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/config/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const imageID = "id-1234"
const profileImageID = "id-1235"
const regionID = "cn_shanghai"

var _ = Describe("Helper", func() {
	DescribeTable("#FindImageForRegion",
		func(profileImages []api.MachineImages, configImages []config.MachineImage, imageName, version, regionID string, expectedImage string) {
			image, err := FindImageForRegion(profileImages, configImages, imageName, version, regionID)

			Expect(image).To(Equal(expectedImage))
			if expectedImage != "" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},

		Entry("list is nil", nil, nil, "ubuntu", "1", regionID, ""),
		Entry("empty list", nil, []config.MachineImage{}, "ubuntu", "1", regionID, ""),
		Entry("entry not found (image does not exist)", nil, makeMachineImages("debian", "1"), "ubuntu", "1", regionID, ""),
		Entry("entry not found (version does not exist)", nil, makeMachineImages("ubuntu", "2"), "ubuntu", "1", regionID, ""),
		Entry("entry not found (region does not exist)", nil, makeMachineImages("ubuntu", "2"), "ubuntu", "2", "cn_beijing", ""),
		Entry("entry", nil, makeMachineImages("ubuntu", "1"), "ubuntu", "1", "cn_shanghai", imageID),

		Entry("profile empty list", []api.MachineImages{}, nil, "ubuntu", "1", regionID, ""),
		Entry("profile entry not found (image does not exist)", makeProfileMachineImages("debian", "1"), nil, "ubuntu", "1", regionID, ""),
		Entry("profile entry not found (version does not exist)", makeProfileMachineImages("ubuntu", "2"), nil, "ubuntu", "1", regionID, ""),
		Entry("profile entry", makeProfileMachineImages("ubuntu", "1"), nil, "ubuntu", "1", regionID, profileImageID),

		Entry("mixed entry not found (image does not exist)", makeProfileMachineImages("debian", "1"), makeMachineImages("debian", "1"), "ubuntu", "1", regionID, ""),
		Entry("mixed entry not found (version does not exist)", makeProfileMachineImages("ubuntu", "2"), makeMachineImages("ubuntu", "2"), "ubuntu", "1", regionID, ""),
		Entry("mixed entry config", makeProfileMachineImages("debian", "1"), makeMachineImages("ubuntu", "1"), "ubuntu", "1", regionID, imageID),
		Entry("mixed entry profile", makeProfileMachineImages("ubuntu", "1"), makeMachineImages("ubuntu", "1"), "ubuntu", "1", "cn_beijing", profileImageID),
		Entry("mixed entry overwrite", makeProfileMachineImages("ubuntu", "1"), makeMachineImages("ubuntu", "1"), "ubuntu", "1", regionID, profileImageID),
	)
})

func makeMachineImages(name, version string) []config.MachineImage {
	return []config.MachineImage{
		{
			Name:    name,
			Version: version,
			Regions: []config.RegionImageMapping{
				{
					Region:  regionID,
					ImageID: imageID,
				},
			},
		},
	}
}

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
