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
	api "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/config"
	. "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/config/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helper", func() {
	DescribeTable("#FindAMIForRegion",
		func(profileImages []api.MachineImages, configImages []config.MachineImage, imageName, version, regionName, expectedAMI string) {
			ami, err := FindAMIForRegion(profileImages, configImages, imageName, version, regionName)

			Expect(ami).To(Equal(expectedAMI))
			if expectedAMI != "" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},

		Entry("list is nil", nil, nil, "ubuntu", "1", "europe", ""),
		Entry("empty list", nil, []config.MachineImage{}, "ubuntu", "1", "europe", ""),
		Entry("entry not found (image does not exist)", nil, makeMachineImages("debian", "1", "europe", "0"), "ubuntu", "1", "europe", ""),
		Entry("entry not found (version does not exist)", nil, makeMachineImages("ubuntu", "2", "europe", "0"), "ubuntu", "1", "europe", ""),
		Entry("entry not found (region does not exist)", nil, makeMachineImages("ubuntu", "1", "asia", "0"), "ubuntu", "1", "europe", ""),
		Entry("entry", nil, makeMachineImages("ubuntu", "1", "europe", "ami-1234"), "ubuntu", "1", "europe", "ami-1234"),

		Entry("profile empty list", []api.MachineImages{}, nil, "ubuntu", "1", "europe", ""),
		Entry("profile entry not found (image does not exist)", makeProfileMachineImages("debian", "1", "europe", "0"), nil, "ubuntu", "1", "europe", ""),
		Entry("profile entry not found (version does not exist)", makeProfileMachineImages("ubuntu", "2", "europe", "0"), nil, "ubuntu", "1", "europe", ""),
		Entry("profile entry", makeProfileMachineImages("ubuntu", "1", "europe", "ami-1234"), nil, "ubuntu", "1", "europe", "ami-1234"),

		Entry("mixed entry not found (image does not exist)", makeProfileMachineImages("debian", "1", "europe", "0"), makeMachineImages("debian", "1", "europe", "1"), "ubuntu", "1", "europe", ""),
		Entry("mixed entry not found (version does not exist)", makeProfileMachineImages("ubuntu", "2", "europe", "0"), makeMachineImages("ubuntu", "2", "europe", "1"), "ubuntu", "1", "europe", ""),
		Entry("mixed entry config", makeProfileMachineImages("debian", "1", "europe", "ami-1234"), makeMachineImages("ubuntu", "1", "europe", "ami-4567"), "ubuntu", "1", "europe", "ami-4567"),
		Entry("mixed entry profile", makeProfileMachineImages("ubuntu", "1", "europe", "ami-1234"), makeMachineImages("ubuntu", "1", "asia", "ami-4567"), "ubuntu", "1", "europe", "ami-1234"),
		Entry("mixed entry overwrite", makeProfileMachineImages("ubuntu", "1", "europe", "ami-1234"), makeMachineImages("ubuntu", "1", "europe", "1"), "ubuntu", "1", "europe", "ami-1234"),
	)
})

func makeMachineImages(name, version, region, ami string) []config.MachineImage {
	var regionAMIMapping []config.RegionAMIMapping
	if len(region) != 0 && len(ami) != 0 {
		regionAMIMapping = append(regionAMIMapping, config.RegionAMIMapping{
			Name: region,
			AMI:  ami,
		})
	}

	return []config.MachineImage{
		{
			Name:    name,
			Version: version,
			Regions: regionAMIMapping,
		},
	}
}

func makeProfileMachineImages(name, version, region, ami string) []api.MachineImages {
	versions := []api.MachineImageVersion{
		api.MachineImageVersion{
			Version: version,
			Regions: []api.RegionAMIMapping{
				{
					Name: region,
					AMI:  ami,
				},
			},
		},
	}

	return []api.MachineImages{
		{
			Name:     name,
			Versions: versions,
		},
	}
}
