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
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"
	. "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helper", func() {
	DescribeTable("#FindInstanceProfileForPurpose",
		func(instanceProfiles []aws.InstanceProfile, purpose string, expectedInstanceProfile *aws.InstanceProfile, expectErr bool) {
			instanceProfile, err := FindInstanceProfileForPurpose(instanceProfiles, purpose)
			expectResults(instanceProfile, expectedInstanceProfile, err, expectErr)
		},

		Entry("list is nil", nil, "foo", nil, true),
		Entry("empty list", []aws.InstanceProfile{}, "foo", nil, true),
		Entry("entry not found", []aws.InstanceProfile{{Name: "bar", Purpose: "baz"}}, "foo", nil, true),
		Entry("entry exists", []aws.InstanceProfile{{Name: "bar", Purpose: "baz"}}, "baz", &aws.InstanceProfile{Name: "bar", Purpose: "baz"}, false),
	)

	DescribeTable("#FindRoleForPurpose",
		func(roles []aws.Role, purpose string, expectedRole *aws.Role, expectErr bool) {
			role, err := FindRoleForPurpose(roles, purpose)
			expectResults(role, expectedRole, err, expectErr)
		},

		Entry("list is nil", nil, "foo", nil, true),
		Entry("empty list", []aws.Role{}, "foo", nil, true),
		Entry("entry not found", []aws.Role{{ARN: "bar", Purpose: "baz"}}, "foo", nil, true),
		Entry("entry exists", []aws.Role{{ARN: "bar", Purpose: "baz"}}, "baz", &aws.Role{ARN: "bar", Purpose: "baz"}, false),
	)

	DescribeTable("#FindSecurityGroupForPurpose",
		func(securityGroups []aws.SecurityGroup, purpose string, expectedSecurityGroup *aws.SecurityGroup, expectErr bool) {
			securityGroup, err := FindSecurityGroupForPurpose(securityGroups, purpose)
			expectResults(securityGroup, expectedSecurityGroup, err, expectErr)
		},

		Entry("list is nil", nil, "foo", nil, true),
		Entry("empty list", []aws.SecurityGroup{}, "foo", nil, true),
		Entry("entry not found", []aws.SecurityGroup{{ID: "bar", Purpose: "baz"}}, "foo", nil, true),
		Entry("entry exists", []aws.SecurityGroup{{ID: "bar", Purpose: "baz"}}, "baz", &aws.SecurityGroup{ID: "bar", Purpose: "baz"}, false),
	)

	DescribeTable("#FindSubnetForPurposeAndZone",
		func(subnets []aws.Subnet, purpose, zone string, expectedSubnet *aws.Subnet, expectErr bool) {
			subnet, err := FindSubnetForPurposeAndZone(subnets, purpose, zone)
			expectResults(subnet, expectedSubnet, err, expectErr)
		},

		Entry("list is nil", nil, "foo", "europe", nil, true),
		Entry("empty list", []aws.Subnet{}, "foo", "europe", nil, true),
		Entry("entry not found (no purpose)", []aws.Subnet{{ID: "bar", Purpose: "baz", Zone: "europe"}}, "foo", "europe", nil, true),
		Entry("entry not found (no zone)", []aws.Subnet{{ID: "bar", Purpose: "baz", Zone: "europe"}}, "foo", "asia", nil, true),
		Entry("entry exists", []aws.Subnet{{ID: "bar", Purpose: "baz", Zone: "europe"}}, "baz", "europe", &aws.Subnet{ID: "bar", Purpose: "baz", Zone: "europe"}, false),
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
