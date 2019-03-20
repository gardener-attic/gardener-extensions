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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

var _ = Describe("Mappers", func() {
	var (
		ctrl          *gomock.Controller
		c             *mockclient.MockClient
		extensionType string
		objectName    string
		resourceName  string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = mockclient.NewMockClient(ctrl)
		extensionType = "certificate-service"
		objectName = "object-abc"
		resourceName = "certificate-service"
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#ObjectNameToExtensionTypeMapper", func() {
		var (
			requestFunc handler.ToRequestsFunc
			object      handler.MapObject
		)

		BeforeEach(func() {
			requestFunc = ObjectNameToExtensionTypeMapper(c, extensionType)
			object = handler.MapObject{
				Meta: &metav1.ObjectMeta{
					Name: objectName,
				},
			}
		})

		It("should find the extension for the passed object", func() {
			geList := extensionsv1alpha1.ExtensionList{
				Items: []extensionsv1alpha1.Extension{
					extensionsv1alpha1.Extension{
						ObjectMeta: metav1.ObjectMeta{
							Name: resourceName,
						},
						Spec: extensionsv1alpha1.ExtensionSpec{
							DefaultSpec: extensionsv1alpha1.DefaultSpec{
								Type: extensionType,
							},
						},
					},
				},
			}

			c.EXPECT().
				List(
					gomock.AssignableToTypeOf(context.TODO()),
					gomock.Eq(client.InNamespace(object.Meta.GetName())),
					gomock.AssignableToTypeOf(&extensionsv1alpha1.ExtensionList{}),
				).
				DoAndReturn(func(_ context.Context, _ *client.ListOptions, actual *extensionsv1alpha1.ExtensionList) error {
					*actual = geList
					return nil
				})

			result := requestFunc(object)

			Expect(result).To(ConsistOf(MatchAllFields(
				Fields{
					"NamespacedName": MatchAllFields(
						Fields{
							"Name":      Equal(resourceName),
							"Namespace": Equal(objectName),
						}),
				},
			)))
		})

		It("should not find the extension for the passed object because extension type is not in the list", func() {
			geList := extensionsv1alpha1.ExtensionList{
				Items: []extensionsv1alpha1.Extension{
					extensionsv1alpha1.Extension{
						ObjectMeta: metav1.ObjectMeta{
							Name: resourceName,
						},
						Spec: extensionsv1alpha1.ExtensionSpec{
							DefaultSpec: extensionsv1alpha1.DefaultSpec{
								Type: "anotherType",
							},
						},
					},
				},
			}

			c.EXPECT().
				List(
					gomock.AssignableToTypeOf(context.TODO()),
					gomock.Eq(client.InNamespace(object.Meta.GetName())),
					gomock.AssignableToTypeOf(&extensionsv1alpha1.ExtensionList{}),
				).
				DoAndReturn(func(_ context.Context, _ *client.ListOptions, actual *extensionsv1alpha1.ExtensionList) error {
					*actual = geList
					return nil
				})

			result := requestFunc(object)

			Expect(result).To(BeEmpty())
		})

		It("should not find the extension for the passed object because extension list is empty", func() {
			c.EXPECT().
				List(
					gomock.AssignableToTypeOf(context.TODO()),
					gomock.Eq(client.InNamespace(object.Meta.GetName())),
					gomock.AssignableToTypeOf(&extensionsv1alpha1.ExtensionList{}),
				).
				DoAndReturn(func(_ context.Context, _ *client.ListOptions, actual *extensionsv1alpha1.ExtensionList) error {
					return nil
				})

			result := requestFunc(object)

			Expect(result).To(BeEmpty())
		})
	})

	Describe("#TypeMapperWithinNamespace", func() {
		var (
			namespace   string
			requestFunc handler.ToRequestsFunc
			object      handler.MapObject
		)

		BeforeEach(func() {
			namespace = "default"
			requestFunc = TypeMapperWithinNamespace(c, extensionType)
			object = handler.MapObject{
				Meta: &metav1.ObjectMeta{
					Name:      objectName,
					Namespace: namespace,
				},
			}
		})

		It("should find the extension for the passed object", func() {
			geList := extensionsv1alpha1.ExtensionList{
				Items: []extensionsv1alpha1.Extension{
					extensionsv1alpha1.Extension{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceName,
							Namespace: namespace,
						},
						Spec: extensionsv1alpha1.ExtensionSpec{
							DefaultSpec: extensionsv1alpha1.DefaultSpec{
								Type: extensionType,
							},
						},
					},
				},
			}

			c.EXPECT().
				List(
					gomock.AssignableToTypeOf(context.TODO()),
					gomock.Eq(client.InNamespace(object.Meta.GetNamespace())),
					gomock.AssignableToTypeOf(&extensionsv1alpha1.ExtensionList{}),
				).
				DoAndReturn(func(_ context.Context, _ *client.ListOptions, actual *extensionsv1alpha1.ExtensionList) error {
					*actual = geList
					return nil
				})

			result := requestFunc(object)

			Expect(result).To(ConsistOf(MatchAllFields(
				Fields{
					"NamespacedName": MatchAllFields(
						Fields{
							"Name":      Equal(resourceName),
							"Namespace": Equal(namespace),
						}),
				},
			)))
		})

		It("should not find the extension for the passed object because extension type is not in the list", func() {
			geList := extensionsv1alpha1.ExtensionList{
				Items: []extensionsv1alpha1.Extension{
					extensionsv1alpha1.Extension{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceName,
							Namespace: namespace,
						},
						Spec: extensionsv1alpha1.ExtensionSpec{
							DefaultSpec: extensionsv1alpha1.DefaultSpec{
								Type: "anotherType",
							},
						},
					},
				},
			}

			c.EXPECT().
				List(
					gomock.AssignableToTypeOf(context.TODO()),
					gomock.Eq(client.InNamespace(object.Meta.GetNamespace())),
					gomock.AssignableToTypeOf(&extensionsv1alpha1.ExtensionList{}),
				).
				DoAndReturn(func(_ context.Context, _ *client.ListOptions, actual *extensionsv1alpha1.ExtensionList) error {
					*actual = geList
					return nil
				})

			result := requestFunc(object)

			Expect(result).To(BeEmpty())
		})

		It("should not find the extension for the passed object because extension list is empty", func() {
			c.EXPECT().
				List(
					gomock.AssignableToTypeOf(context.TODO()),
					gomock.Eq(client.InNamespace(object.Meta.GetNamespace())),
					gomock.AssignableToTypeOf(&extensionsv1alpha1.ExtensionList{}),
				).
				DoAndReturn(func(_ context.Context, _ *client.ListOptions, actual *extensionsv1alpha1.ExtensionList) error {
					return nil
				})

			result := requestFunc(object)

			Expect(result).To(BeEmpty())
		})
	})
})
