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
	apisconfig "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ChartValues", func() {

	var (
		certmanagementConfig *apisconfig.ConfigurationSpec
		route53Provider      apisconfig.DNSProviderConfig
		cloudDNSProvider     apisconfig.DNSProviderConfig
		namespaceRef         string
		namespaceUID         types.UID
	)

	BeforeEach(func() {
		route53Provider = &apisconfig.Route53{
			Domains:         []string{"example.com"},
			Name:            "route53",
			Region:          "eu-west-1",
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretAccessKey",
		}

		cloudDNSProvider = &apisconfig.CloudDNS{
			Domains:        []string{"example.org"},
			Name:           "clouddns",
			Project:        "project-id",
			ServiceAccount: "svcJson",
		}

		certmanagementConfig = &apisconfig.ConfigurationSpec{
			IssuerName: "issuer",
			ACME: apisconfig.ACME{
				Email:  "operator@gardener.cloud",
				Server: "https://letsencrypt.org/v2/staging",
			},
			Providers: apisconfig.DNSProviders{
				Route53: []apisconfig.Route53{
					*route53Provider.(*apisconfig.Route53),
				},
				CloudDNS: []apisconfig.CloudDNS{
					*cloudDNSProvider.(*apisconfig.CloudDNS),
				},
			},
		}

		namespaceRef = "extension-certificate-service"
		namespaceUID = "123-abc"
	})

	Describe("#CreateDNSProviderValues", func() {
		It("should compute route53 values correctly", func() {
			values, err := CreateDNSProviderValues([]apisconfig.DNSProviderConfig{route53Provider})

			expectedValues := map[string]interface{}{
				"type":        apisconfig.Route53Provider,
				"region":      "eu-west-1",
				"accessKeyID": "accessKey",
				"accessKey":   "secretAccessKey",
				"name":        "route53",
			}

			Expect(err).To(BeNil())
			Expect(values[0]).To(Equal(expectedValues))
		})

		It("should compute cloud DNS values correctly", func() {
			values, err := CreateDNSProviderValues([]apisconfig.DNSProviderConfig{cloudDNSProvider})

			expectedValues := map[string]interface{}{
				"type":      apisconfig.CloudDNSProvider,
				"project":   "project-id",
				"name":      "clouddns",
				"accessKey": "svcJson",
			}

			Expect(err).To(BeNil())
			Expect(values[0]).To(Equal(expectedValues))
		})
	})

	Describe("#CreateCertServiceValues", func() {
		It("should compute values correctly", func() {
			values, err := CreateCertServiceValues(*certmanagementConfig, namespaceRef, namespaceUID)

			expectedValues := map[string]interface{}{
				"clusterissuer": map[string]interface{}{
					"name": certmanagementConfig.IssuerName,
					"acme": map[string]interface{}{
						"email":  certmanagementConfig.ACME.Email,
						"server": certmanagementConfig.ACME.Server,
						"letsEncrypt": map[string]interface{}{
							"name": "lets-encrypt",
							"key":  "",
						},
						"dns01": map[string]interface{}{
							"providers": []map[string]interface{}{
								{
									"type":        apisconfig.Route53Provider,
									"region":      "eu-west-1",
									"accessKeyID": "accessKey",
									"accessKey":   "secretAccessKey",
									"name":        "route53",
								},
								{
									"type":      apisconfig.CloudDNSProvider,
									"project":   "project-id",
									"accessKey": "svcJson",
									"name":      "clouddns",
								},
							},
						},
					},
				},
				"namespaceRef": map[string]interface{}{
					"namespace": namespaceRef,
					"uid":       namespaceUID,
				},
			}

			Expect(err).To(BeNil())
			Expect(values).To(Equal(expectedValues))
		})
	})

})
