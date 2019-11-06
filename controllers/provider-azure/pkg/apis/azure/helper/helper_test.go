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
	"github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/azure"
	. "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/azure/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helper", func() {
	var (
		purpose      azure.Purpose = "foo"
		purposeWrong azure.Purpose = "baz"
		urn          string        = "publisher:offer:sku:version"
	)

	DescribeTable("#FindSubnetByPurpose",
		func(subnets []azure.Subnet, purpose azure.Purpose, expectedSubnet *azure.Subnet, expectErr bool) {
			subnet, err := FindSubnetByPurpose(subnets, purpose)
			expectResults(subnet, expectedSubnet, err, expectErr)
		},

		Entry("list is nil", nil, purpose, nil, true),
		Entry("empty list", []azure.Subnet{}, purpose, nil, true),
		Entry("entry not found", []azure.Subnet{{Name: "bar", Purpose: purposeWrong}}, purpose, nil, true),
		Entry("entry exists", []azure.Subnet{{Name: "bar", Purpose: purpose}}, purpose, &azure.Subnet{Name: "bar", Purpose: purpose}, false),
	)

	DescribeTable("#FindSecurityGroupByPurpose",
		func(securityGroups []azure.SecurityGroup, purpose azure.Purpose, expectedSecurityGroup *azure.SecurityGroup, expectErr bool) {
			securityGroup, err := FindSecurityGroupByPurpose(securityGroups, purpose)
			expectResults(securityGroup, expectedSecurityGroup, err, expectErr)
		},

		Entry("list is nil", nil, purpose, nil, true),
		Entry("empty list", []azure.SecurityGroup{}, purpose, nil, true),
		Entry("entry not found", []azure.SecurityGroup{{Name: "bar", Purpose: purposeWrong}}, purpose, nil, true),
		Entry("entry exists", []azure.SecurityGroup{{Name: "bar", Purpose: purpose}}, purpose, &azure.SecurityGroup{Name: "bar", Purpose: purpose}, false),
	)

	DescribeTable("#FindRouteTableByPurpose",
		func(routeTables []azure.RouteTable, purpose azure.Purpose, expectedRouteTable *azure.RouteTable, expectErr bool) {
			routeTable, err := FindRouteTableByPurpose(routeTables, purpose)
			expectResults(routeTable, expectedRouteTable, err, expectErr)
		},

		Entry("list is nil", nil, purpose, nil, true),
		Entry("empty list", []azure.RouteTable{}, purpose, nil, true),
		Entry("entry not found", []azure.RouteTable{{Name: "bar", Purpose: purposeWrong}}, purpose, nil, true),
		Entry("entry exists", []azure.RouteTable{{Name: "bar", Purpose: purpose}}, purpose, &azure.RouteTable{Name: "bar", Purpose: purpose}, false),
	)

	DescribeTable("#FindAvailabilitySetByPurpose",
		func(availabilitySets []azure.AvailabilitySet, purpose azure.Purpose, expectedAvailabilitySet *azure.AvailabilitySet, expectErr bool) {
			availabilitySet, err := FindAvailabilitySetByPurpose(availabilitySets, purpose)
			expectResults(availabilitySet, expectedAvailabilitySet, err, expectErr)
		},

		Entry("list is nil", nil, purpose, nil, true),
		Entry("empty list", []azure.AvailabilitySet{}, purpose, nil, true),
		Entry("entry not found", []azure.AvailabilitySet{{ID: "bar", Purpose: purposeWrong}}, purpose, nil, true),
		Entry("entry exists", []azure.AvailabilitySet{{ID: "bar", Purpose: purpose}}, purpose, &azure.AvailabilitySet{ID: "bar", Purpose: purpose}, false),
	)

	DescribeTable("#FindMachineImage",
		func(machineImages []azure.MachineImage, name, version string, expectedMachineImage *azure.MachineImage, expectErr bool) {
			machineImage, err := FindMachineImage(machineImages, name, version)
			expectResults(machineImage, expectedMachineImage, err, expectErr)
		},

		Entry("list is nil", nil, "foo", "1.2.3", nil, true),
		Entry("empty list", []azure.MachineImage{}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no name)", []azure.MachineImage{{Name: "bar", Version: "1.2.3", URN: &urn}}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no version)", []azure.MachineImage{{Name: "bar", Version: "1.2.3", URN: &urn}}, "bar", "1.2.4", nil, true),
		Entry("entry exists", []azure.MachineImage{{Name: "bar", Version: "1.2.3", URN: &urn}}, "bar", "1.2.3", &azure.MachineImage{Name: "bar", Version: "1.2.3", URN: &urn}, false),
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
