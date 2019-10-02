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

package helper_test

import (
	api "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere"
	. "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const somePath = "/some/somePath"
const someGuestId = "linux64"

var _ = Describe("Helper", func() {
	var barMachine = api.MachineImage{Name: "bar", Version: "1.2.3", Path: "/x/y"}

	DescribeTable("#FindMachineImage",
		func(configImages []api.MachineImage, name, version string, expectedMachineImage *api.MachineImage, expectErr bool) {
			machineImage, err := FindMachineImage(configImages, name, version)
			expectResults(machineImage, expectedMachineImage, err, expectErr)
		},

		Entry("list is nil", nil, "foo", "1.2.3", nil, true),
		Entry("empty list", []api.MachineImage{}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no name)", []api.MachineImage{}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no version)", []api.MachineImage{barMachine}, barMachine.Name, "1.2.4", nil, true),
		Entry("entry exists", []api.MachineImage{barMachine}, barMachine.Name, barMachine.Version, &barMachine, false),
	)

	DescribeTable("#FindImage",
		func(profileImages []api.MachineImages, imageName, version string, expectedPath string) {
			path, guestId, err := FindImage(profileImages, imageName, version)

			Expect(path).To(Equal(expectedPath))
			if expectedPath != "" {
				Expect(guestId).To(Equal(someGuestId))
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(guestId).To(Equal(""))
				Expect(err).To(HaveOccurred())
			}
		},

		Entry("list is nil", nil, "ubuntu", "1", ""),
		Entry("profile empty list", []api.MachineImages{}, "ubuntu", "1", ""),
		Entry("profile entry not found (image does not exist)", makeProfileMachineImages("debian", "1"), "ubuntu", "1", ""),
		Entry("profile entry not found (version does not exist)", makeProfileMachineImages("ubuntu", "2"), "ubuntu", "1", ""),
		Entry("profile entry", makeProfileMachineImages("ubuntu", "1"), "ubuntu", "1", somePath),
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

func makeProfileMachineImages(name, version string) []api.MachineImages {
	return []api.MachineImages{
		{
			Name: name,
			Versions: []api.MachineImageVersion{
				{
					Version: version,
					Path:    somePath,
					GuestID: someGuestId,
				},
			},
		},
	}
}
