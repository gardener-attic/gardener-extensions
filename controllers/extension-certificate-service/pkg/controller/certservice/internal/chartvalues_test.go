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

package internal

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/extensions/table"

	apisconfig "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config"
)

var _ = Describe("ChartValues", func() {
	var (
		provider     *apisconfig.Route53
		providerName = "dnsProvider"
	)

	BeforeEach(func() {
		provider = &apisconfig.Route53{
			Domains: []string{},
			Name:    providerName,
		}
	})

	DescribeTable("CreateDNSProviderValue",
		func(providerDomains []string, shootDomain string, expectation map[string]string) {
			provider.Domains = providerDomains
			value := CreateDNSProviderValue(provider, shootDomain)
			Expect(value).To(Equal(expectation))
		},
		Entry("exact match one domain", []string{"example.com"}, "example.com", valueMap(providerName, "example.com")),
		Entry("exact match multiple domains", []string{"foo.bar", "example.com", "bar.foo"}, "example.com", valueMap(providerName, "example.com")),
		Entry("subdomain match", []string{"foo.bar", "example.com"}, "shoot.project.example.com", valueMap(providerName, "shoot.project.example.com")),
		Entry("no match", []string{"foo.bar", "example.com"}, "shoot.aexample.com", nil),
	)
})

func valueMap(name, domain string) map[string]string {
	return map[string]string{
		"provider": name,
		"domain":   domain,
	}
}
