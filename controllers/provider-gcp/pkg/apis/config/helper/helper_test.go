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
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/config"
	. "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/config/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const image = "project/path/to/image"

var _ = Describe("Helper", func() {
	DescribeTable("#FindImage",
		func(machineImages []config.MachineImage, imageName, version string, expectedImage string) {
			image, err := FindImage(machineImages, imageName, version)

			Expect(image).To(Equal(expectedImage))
			if expectedImage != "" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},

		Entry("list is nil", nil, "ubuntu", "1", ""),
		Entry("empty list", []config.MachineImage{}, "ubuntu", "1", ""),
		Entry("entry not found (image does not exist)", makeMachineImages("debian", "1"), "ubuntu", "1", ""),
		Entry("entry not found (version does not exist)", makeMachineImages("ubuntu", "2"), "ubuntu", "1", ""),
		Entry("entry", makeMachineImages("ubuntu", "1"), "ubuntu", "1", image),
	)
})

func makeMachineImages(name, version string) []config.MachineImage {
	return []config.MachineImage{
		{
			Name:    name,
			Version: version,
			Image:   image,
		},
	}
}
