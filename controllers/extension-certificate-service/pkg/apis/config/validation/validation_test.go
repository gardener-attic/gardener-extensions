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

package validation

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTypeValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Certificate Service Types Validation Suite")
}

var _ = Describe("Validation", func() {
	var (
		certmanagementConfig        *config.Configuration
		invalidCertmanagementConfig *config.Configuration
		route53Provider             *config.Route53
		invalidRoute53Provider      *config.Route53
		cloudDNSProvider            *config.CloudDNS
		invalidCloudDNSProvider     *config.CloudDNS
	)

	BeforeEach(func() {
		route53Provider = &config.Route53{
			Domains:         []string{"example.com"},
			Name:            "route53",
			Region:          "eu-west-1",
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretAccessKey",
		}

		invalidRoute53Provider = &config.Route53{}

		cloudDNSProvider = &config.CloudDNS{
			Domains:        []string{"example.org"},
			Name:           "clouddns",
			Project:        "project-id",
			ServiceAccount: "svcJson",
		}

		invalidCloudDNSProvider = &config.CloudDNS{}

		certmanagementConfig = &config.Configuration{
			Spec: config.ConfigurationSpec{
				LifecycleSync:     metav1.Duration{Duration: 1 * time.Hour},
				ServiceSync:       metav1.Duration{Duration: 5 * time.Minute},
				IssuerName:        "issuer",
				ResourceNamespace: "garden",
				NamespaceRef:      "extension-controller",
				ACME: config.ACME{
					Email:  "operator@gardener.cloud",
					Server: "https://letsencrypt.org/v2/staging",
				},
			},
		}

		invalidCertmanagementConfig = &config.Configuration{}
	})

	Describe("#ValidateConfiguration", func() {
		It("should validate configuration w/o errors", func() {
			certmanagementConfig.Spec.Providers.Route53 = []config.Route53{*route53Provider}
			certmanagementConfig.Spec.Providers.CloudDNS = []config.CloudDNS{*cloudDNSProvider}
			errs := ValidateConfiguration(certmanagementConfig)

			Expect(errs).To(BeEmpty())
		})

		It("should exit validation w/ errors", func() {
			invalidCertmanagementConfig.Spec.Providers.Route53 = []config.Route53{*invalidRoute53Provider}
			invalidCertmanagementConfig.Spec.Providers.CloudDNS = []config.CloudDNS{*invalidCloudDNSProvider}
			errs := ValidateConfiguration(invalidCertmanagementConfig)

			Expect(errs).ToNot(BeEmpty())
			Expect(errs).Should(
				ConsistOf(
					// Missing IssuerName
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeRequired)})),
					// Missing ResourceNamespace
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeRequired)})),
					// Missing NamespaceRef
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeRequired)})),
					// Invalid LifecycleSync
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeInvalid)})),
					// Invalid ServiceSync
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeInvalid)})),

					//ACME
					// Invalid Server
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeInvalid)})),
					// Invalid Email
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeInvalid)})),

					// Route53
					// Missing Name
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeRequired)})),
					// Missing Region
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeRequired)})),
					// Missing AccessKeyID
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeRequired)})),
					// Missing SecretAccessKey
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeRequired)})),

					// CloudDNS
					// Missing Name
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeRequired)})),
					// Missing Project
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeRequired)})),
					// Missing ServiceAccount
					PointTo(MatchFields(IgnoreExtras, Fields{"Type": Equal(field.ErrorTypeRequired)})),
				),
			)
		})
	})
})
