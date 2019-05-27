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

package client

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client")
}

var _ = Describe("Client Suite", func() {
	Describe("#ReadSecretCredentials", func() {
		It("should correctly read the credentials", func() {
			var (
				accessKeyID     = "accessKeyID"
				accessKeySecret = "accessKeySecret"
			)

			creds, err := ReadSecretCredentials(&corev1.Secret{
				Data: map[string][]byte{
					AccessKeyIDField:     []byte(accessKeyID),
					AccessKeySecretField: []byte(accessKeySecret),
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(creds).To(Equal(&Credentials{
				AccessKeyID:     accessKeyID,
				AccessKeySecret: accessKeySecret,
			}))
		})

		It("should error if the data section is nil", func() {
			_, err := ReadSecretCredentials(&corev1.Secret{})
			Expect(err).To(HaveOccurred())
		})

		It("should error if access key id is missing", func() {
			var (
				accessKeySecret = "accessKeySecret"
			)

			_, err := ReadSecretCredentials(&corev1.Secret{
				Data: map[string][]byte{
					AccessKeySecretField: []byte(accessKeySecret),
				},
			})

			Expect(err).To(HaveOccurred())
		})

		It("should error if access key secret is missing", func() {
			var (
				accessKeyID = "accessKeyID"
			)

			_, err := ReadSecretCredentials(&corev1.Secret{
				Data: map[string][]byte{
					AccessKeyIDField: []byte(accessKeyID),
				},
			})

			Expect(err).To(HaveOccurred())
		})
	})
})
