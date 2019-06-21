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

package resourcemanager

import (
	"context"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

	resourcemanagerv1alpha1 "github.com/gardener/gardener-resource-manager/pkg/apis/resources/v1alpha1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resource Manager Test Suite")
}

var _ = Describe("Resource Manager", func() {
	var (
		ctrl *gomock.Controller
		c    *mockclient.MockClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = mockclient.NewMockClient(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Secrets", func() {
		var ctx = context.TODO()

		It("should create a managed secret ", func() {
			var (
				secretName      = "foo"
				secretNamespace = "bar"

				secretData = map[string][]byte{
					"foo": []byte("bar"),
				}

				secretMeta = metav1.ObjectMeta{
					Name:      secretName,
					Namespace: secretNamespace,
				}
				expectedSecret = corev1.Secret{
					ObjectMeta: secretMeta,
					Data:       secretData,
					Type:       corev1.SecretTypeOpaque,
				}
			)

			managedSecret := NewSecret(c).WithNamespacedName(types.NamespacedName{Name: secretName, Namespace: secretNamespace}).WithKeyValues(secretData)
			Expect(managedSecret.secret).To(Equal(expectedSecret))

			c.EXPECT().Get(ctx, kutil.Key(secretNamespace, secretName), gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
				return apierrors.NewNotFound(corev1.Resource("secrets"), secretName)
			})

			c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, secret *corev1.Secret) error {
				secret.ObjectMeta = secretMeta
				secret.Data = secretData
				return nil
			})

			err := managedSecret.Reconcile(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should create a managed resource ", func() {
			var (
				managedResourceName      = "foo"
				managedResourceNamespace = "bar"

				managedResourceMeta = metav1.ObjectMeta{
					Name:      managedResourceName,
					Namespace: managedResourceNamespace,
				}

				secretRefs = []corev1.LocalObjectReference{
					{Name: "test1"},
					{Name: "test2"},
				}

				injectedLabels = map[string]string{
					"shoot.gardener.cloud/no-cleanup": "true",
				}

				expectedManagedResource = resourcemanagerv1alpha1.ManagedResource{
					ObjectMeta: managedResourceMeta,
					Spec: resourcemanagerv1alpha1.ManagedResourceSpec{
						SecretRefs:   secretRefs,
						InjectLabels: injectedLabels,
					},
				}
			)

			managedResource := NewManagedResource(c).WithNamespacedName(types.NamespacedName{Name: managedResourceName, Namespace: managedResourceNamespace}).WithSecretRefs(secretRefs).WithInjectedLabels(injectedLabels)
			Expect(managedResource.resource).To(Equal(expectedManagedResource))

			c.EXPECT().Get(ctx, kutil.Key(managedResourceNamespace, managedResourceName), gomock.AssignableToTypeOf(&resourcemanagerv1alpha1.ManagedResource{})).DoAndReturn(func(_ context.Context, _ client.ObjectKey, ms *resourcemanagerv1alpha1.ManagedResource) error {
				return apierrors.NewNotFound(corev1.Resource("managedresources"), managedResourceName)
			})

			c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&resourcemanagerv1alpha1.ManagedResource{})).Do(func(_ context.Context, ms *resourcemanagerv1alpha1.ManagedResource) error {
				return nil
			})

			err := managedResource.Reconcile(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
