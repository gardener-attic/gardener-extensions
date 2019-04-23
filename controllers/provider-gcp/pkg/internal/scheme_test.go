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

package internal_test

import (
	gcpv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp/v1alpha1"
	. "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal"
	testinfra "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/test/infrastructure"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Scheme Suite", func() {
	var (
		config    gcpv1alpha1.InfrastructureConfig
		rawConfig []byte
		infra     extensionsv1alpha1.Infrastructure
	)
	BeforeEach(func() {
		config = gcpv1alpha1.InfrastructureConfig{
			Networks: gcpv1alpha1.NetworkConfig{
				Worker: gardencorev1alpha1.CIDR("192.168.0.0/16"),
			},
		}
		var err error
		rawConfig, err = testinfra.MkRawInfrastructureConfig(&config)
		Expect(err).NotTo(HaveOccurred())
		infra = extensionsv1alpha1.Infrastructure{
			Spec: extensionsv1alpha1.InfrastructureSpec{
				ProviderConfig: &runtime.RawExtension{
					Raw: rawConfig,
				},
			},
		}
	})

	Describe("#InfrastructureConfigFromInfrastructure", func() {
		It("should correctly extract the infrastructure config from the infrastructure", func() {
			actual, err := InfrastructureConfigFromInfrastructure(&infra)

			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(&config))
		})
	})
})
