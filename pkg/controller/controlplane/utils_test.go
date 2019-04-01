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

package controlplane

import (
	"context"
	"testing"

	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace = "default"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controlplane Suite")
}

var _ = Describe("Utils", func() {
	var (
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#DNSNamesForService", func() {
		It("should return all expected DNS names for the given service name and namespace", func() {
			Expect(DNSNamesForService("test", namespace)).To(Equal([]string{
				"test",
				"test." + namespace,
				"test." + namespace + ".svc",
				"test." + namespace + ".svc.cluster.local",
			}))
		})
	})

	Describe("#ComputeCheckums", func() {
		var (
			secrets = map[string]*corev1.Secret{
				"test-secret": getSecret(namespace, "test-secret", map[string][]byte{"foo": []byte("bar")}),
			}
			cms = map[string]*corev1.ConfigMap{
				"test-config": getConfigMap(namespace, "test-config", map[string]string{"abc": "xyz"}),
			}
			existingSecret = getSecret(namespace, "existing-secret", map[string][]byte{"123": []byte("456")})
			existingConfig = getConfigMap(namespace, "existing-config", map[string]string{"right": "wrong"})
		)
		It("should compute all checksums for the given secrets and configmpas", func() {

			// Create mock client
			c := mockclient.NewMockClient(ctrl)
			c.EXPECT().Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: "existing-secret"}, &corev1.Secret{}).DoAndReturn(
				func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
					*obj.(*corev1.Secret) = *existingSecret
					return nil
				},
			)
			c.EXPECT().Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: "existing-config"}, &corev1.ConfigMap{}).DoAndReturn(
				func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
					*obj.(*corev1.ConfigMap) = *existingConfig
					return nil
				},
			)

			checksums, err := ComputeChecksums(context.TODO(), c, secrets, cms, []string{"existing-secret"}, []string{"existing-config"}, namespace)
			Expect(err).To(Not(HaveOccurred()))
			Expect(checksums).To(Equal(map[string]string{
				"test-secret":     "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
				"test-config":     "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606",
				"existing-secret": "cc5a179ccb1ea04bbf07089102398ca852fc6ac4e5c076ab4451e191102b240f",
				"existing-config": "66e322807a642afda73ea4696ff94362a3b402fcdbe4c5676543478c63f26bb9",
			}))
		})
	})
})

func getSecret(namespace, name string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}

func getConfigMap(namespace, name string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}
