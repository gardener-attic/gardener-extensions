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
	azurev1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/azure/v1alpha1"
	. "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/azure/v1alpha1/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helper", func() {
	DescribeTable("#FindDomainCountByRegion",
		func(domainCounts []azurev1alpha1.DomainCount, region string, expectedCount int, expectErr bool) {
			count, err := FindDomainCountByRegion(domainCounts, region)
			expectResults(count, expectedCount, err, expectErr)
		},

		Entry("list is nil", nil, "foo", 0, true),
		Entry("empty list", []azurev1alpha1.DomainCount{}, "foo", 0, true),
		Entry("entry not found", []azurev1alpha1.DomainCount{{Region: "bar", Count: 1}}, "foo", 0, true),
		Entry("entry exists", []azurev1alpha1.DomainCount{{Region: "bar", Count: 1}}, "bar", 1, false),
	)
})

func expectResults(result, expected interface{}, err error, expectErr bool) {
	if !expectErr {
		Expect(result).To(Equal(expected))
		Expect(err).NotTo(HaveOccurred())
	} else {
		Expect(result).To(BeZero())
		Expect(err).To(HaveOccurred())
	}
}
