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

package handler

import (
	"context"
	extensionspredicate "github.com/gardener/gardener-extensions/pkg/predicate"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/onsi/gomega/gstruct"

	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handler Suite")
}

var _ = Describe("Controller Mapper", func() {
	var (
		ctrl *gomock.Controller
		c    *mockclient.MockClient

		extensionType  string
		objectName     string
		resourceName   string
		newObjListFunc = func() runtime.Object { return &extensionsv1alpha1.ExtensionList{} }
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

	Describe("#TypeMapperWithinNamespace", func() {
		var (
			namespace   string
			requestFunc handler.ToRequestsFunc
			object      handler.MapObject
		)

		BeforeEach(func() {
			namespace = "default"
			requestFunc = MapperWithinNamespace(c, newObjListFunc, nil)
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
					{
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
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      resourceName,
							Namespace: namespace,
						},
						Spec: extensionsv1alpha1.ExtensionSpec{
							DefaultSpec: extensionsv1alpha1.DefaultSpec{
								Type: "bar",
							},
						},
					},
				},
			}
			requestFunc := MapperWithinNamespace(c, newObjListFunc, []predicate.Predicate{extensionspredicate.HasType("foo")})

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

	Describe("#ClusterToObjectMapper", func() {
		var (
			resourceName = "infra"
			namespace    = "shoot"

			newObjListFunc = func() runtime.Object { return &extensionsv1alpha1.InfrastructureList{} }
		)

		It("should find all objects for the passed cluster", func() {
			mapper := ClusterToObjectMapper(c, newObjListFunc, nil)

			c.EXPECT().
				List(
					gomock.AssignableToTypeOf(context.TODO()),
					gomock.Eq(client.InNamespace(namespace)),
					gomock.AssignableToTypeOf(&extensionsv1alpha1.InfrastructureList{}),
				).
				DoAndReturn(func(_ context.Context, _ *client.ListOptions, actual *extensionsv1alpha1.InfrastructureList) error {
					*actual = extensionsv1alpha1.InfrastructureList{
						Items: []extensionsv1alpha1.Infrastructure{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      resourceName,
									Namespace: namespace,
								},
							},
						},
					}
					return nil
				})

			result := mapper.Map(handler.MapObject{
				Object: &extensionsv1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: namespace,
					},
				},
			})

			Expect(result).To(ConsistOf(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				},
			}))
		})

		It("should find no objects for the passed cluster because predicates do not match", func() {
			var (
				predicates = []predicate.Predicate{
					predicate.Funcs{
						GenericFunc: func(event event.GenericEvent) bool {
							return false
						},
					},
				}
				mapper = ClusterToObjectMapper(c, newObjListFunc, predicates)
			)

			c.EXPECT().
				List(
					gomock.AssignableToTypeOf(context.TODO()),
					gomock.Eq(client.InNamespace(namespace)),
					gomock.AssignableToTypeOf(&extensionsv1alpha1.InfrastructureList{}),
				).
				DoAndReturn(func(_ context.Context, _ *client.ListOptions, actual *extensionsv1alpha1.InfrastructureList) error {
					*actual = extensionsv1alpha1.InfrastructureList{
						Items: []extensionsv1alpha1.Infrastructure{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      resourceName,
									Namespace: namespace,
								},
							},
						},
					}
					return nil
				})

			result := mapper.Map(handler.MapObject{
				Object: &extensionsv1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: namespace,
					},
				},
			})

			Expect(result).To(BeEmpty())
		})

		It("should find no objects because list is empty", func() {
			mapper := ClusterToObjectMapper(c, newObjListFunc, nil)

			c.EXPECT().
				List(
					gomock.AssignableToTypeOf(context.TODO()),
					gomock.Eq(client.InNamespace(namespace)),
					gomock.AssignableToTypeOf(&extensionsv1alpha1.InfrastructureList{}),
				).
				DoAndReturn(func(_ context.Context, _ *client.ListOptions, actual *extensionsv1alpha1.InfrastructureList) error {
					*actual = extensionsv1alpha1.InfrastructureList{}
					return nil
				})

			result := mapper.Map(handler.MapObject{
				Object: &extensionsv1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: namespace,
					},
				},
			})

			Expect(result).To(BeEmpty())
		})

		It("should find no objects because the passed object is no cluster", func() {
			mapper := ClusterToObjectMapper(c, newObjListFunc, nil)
			result := mapper.Map(handler.MapObject{
				Object: &extensionsv1alpha1.Infrastructure{},
			})

			Expect(result).To(BeNil())
		})
	})
})
