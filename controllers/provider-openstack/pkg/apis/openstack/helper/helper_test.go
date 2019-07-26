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
	"github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack"
	. "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helper", func() {
	var (
		purpose      openstack.Purpose = "foo"
		purposeWrong openstack.Purpose = "baz"
	)

	DescribeTable("#FindSubnetByPurpose",
		func(subnets []openstack.Subnet, purpose openstack.Purpose, expectedSubnet *openstack.Subnet, expectErr bool) {
			subnet, err := FindSubnetByPurpose(subnets, purpose)
			expectResults(subnet, expectedSubnet, err, expectErr)
		},

		Entry("list is nil", nil, purpose, nil, true),
		Entry("empty list", []openstack.Subnet{}, purpose, nil, true),
		Entry("entry not found", []openstack.Subnet{{ID: "bar", Purpose: purposeWrong}}, purpose, nil, true),
		Entry("entry exists", []openstack.Subnet{{ID: "bar", Purpose: purpose}}, purpose, &openstack.Subnet{ID: "bar", Purpose: purpose}, false),
	)

	DescribeTable("#FindSecurityGroupByPurpose",
		func(securityGroups []openstack.SecurityGroup, purpose openstack.Purpose, expectedSecurityGroup *openstack.SecurityGroup, expectErr bool) {
			securityGroup, err := FindSecurityGroupByPurpose(securityGroups, purpose)
			expectResults(securityGroup, expectedSecurityGroup, err, expectErr)
		},

		Entry("list is nil", nil, purpose, nil, true),
		Entry("empty list", []openstack.SecurityGroup{}, purpose, nil, true),
		Entry("entry not found", []openstack.SecurityGroup{{Name: "bar", Purpose: purposeWrong}}, purpose, nil, true),
		Entry("entry exists", []openstack.SecurityGroup{{Name: "bar", Purpose: purpose}}, purpose, &openstack.SecurityGroup{Name: "bar", Purpose: purpose}, false),
	)

	DescribeTable("#FindMachineImage",
		func(machineImages []openstack.MachineImage, name, version, cloudProfile string, expectedMachineImage *openstack.MachineImage, expectErr bool) {
			machineImage, err := FindMachineImage(machineImages, name, version, cloudProfile)
			expectResults(machineImage, expectedMachineImage, err, expectErr)
		},

		Entry("list is nil", nil, "foo", "1.2.3", "openstack1", nil, true),
		Entry("empty list", []openstack.MachineImage{}, "foo", "1.2.3", "openstack1", nil, true),
		Entry("entry not found (no name)", []openstack.MachineImage{{Name: "bar", Version: "1.2.3", CloudProfile: "openstack1"}}, "foo", "1.2.3", "openstack1", nil, true),
		Entry("entry not found (no version)", []openstack.MachineImage{{Name: "bar", Version: "1.2.3", CloudProfile: "openstack1"}}, "foo", "1.2.3", "openstack2", nil, true),
		Entry("entry not found (no cloud profile)", []openstack.MachineImage{{Name: "bar", Version: "1.2.3", CloudProfile: "openstack1"}}, "bar", "1.2.3", "openstack2", nil, true),
		Entry("entry exists", []openstack.MachineImage{{Name: "bar", Version: "1.2.3", CloudProfile: "openstack1"}}, "bar", "1.2.3", "openstack1", &openstack.MachineImage{Name: "bar", Version: "1.2.3", CloudProfile: "openstack1"}, false),
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
