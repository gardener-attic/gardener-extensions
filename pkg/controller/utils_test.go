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
	"errors"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockmanager "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/manager"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewAppender(v *[]int, i int, mgr manager.Manager, err error) func(manager.Manager) error {
	return func(actual manager.Manager) error {
		*v = append(*v, i)
		Expect(actual).To(Equal(mgr))
		return err
	}
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

	Context("AddToManagerBuilder", func() {
		Describe("#NewAddToManagerBuilder", func() {
			It("should initialize a builder and register the given functions", func() {
				var ids []int
				mgr := mockmanager.NewMockManager(ctrl)

				f1 := NewAppender(&ids, 1, mgr, nil)
				f2 := NewAppender(&ids, 2, mgr, nil)

				builder := NewAddToManagerBuilder(f1, f2)

				Expect(builder[0](mgr)).NotTo(HaveOccurred())
				Expect(builder[1](mgr)).NotTo(HaveOccurred())
				Expect(ids).To(Equal([]int{1, 2}))
			})
		})

		Describe("#Register", func() {
			It("should register the new functions in the builder", func() {
				var ids []int
				mgr := mockmanager.NewMockManager(ctrl)

				f1 := NewAppender(&ids, 1, mgr, nil)
				f2 := NewAppender(&ids, 2, mgr, nil)
				f3 := NewAppender(&ids, 3, mgr, nil)

				builder := NewAddToManagerBuilder(f1)

				builder.Register(f2, f3)

				Expect(builder[0](mgr)).NotTo(HaveOccurred())
				Expect(builder[1](mgr)).NotTo(HaveOccurred())
				Expect(builder[2](mgr)).NotTo(HaveOccurred())
				Expect(ids).To(Equal([]int{1, 2, 3}))
			})
		})

		Describe("#AddToManager", func() {
			It("should call the functions in the correct sequence", func() {
				var ids []int
				mgr := mockmanager.NewMockManager(ctrl)

				f1 := NewAppender(&ids, 1, mgr, nil)
				f2 := NewAppender(&ids, 2, mgr, nil)
				f3 := NewAppender(&ids, 3, mgr, nil)

				builder := NewAddToManagerBuilder(f1, f2, f3)
				Expect(builder.AddToManager(mgr)).NotTo(HaveOccurred())
				Expect(ids).To(Equal([]int{1, 2, 3}))
			})

			It("should exit on the first error and return it", func() {
				var ids []int
				mgr := mockmanager.NewMockManager(ctrl)
				err := errors.New("error")

				f1 := NewAppender(&ids, 1, mgr, nil)
				f2 := NewAppender(&ids, 2, mgr, err)
				f3 := NewAppender(&ids, 3, mgr, nil)

				builder := NewAddToManagerBuilder(f1, f2, f3)
				Expect(builder.AddToManager(mgr)).To(BeIdenticalTo(err))
				Expect(ids).To(Equal([]int{1, 2}))
			})
		})
	})

	Context("Finalizers", func() {
		const finalizerName = "foo"
		Describe("#EnsureFinalizer", func() {

			It("should add a finalizer if none was present", func() {
				secret := &corev1.Secret{}
				c := mockclient.NewMockClient(ctrl)
				c.EXPECT().Update(nil, secret)

				err := EnsureFinalizer(nil, c, finalizerName, secret)
				Expect(err).NotTo(HaveOccurred())
				Expect(secret.GetFinalizers()).To(Equal([]string{finalizerName}))
			})

			It("should not add a finalizer if it was already present", func() {
				secret := &corev1.Secret{}
				secret.SetFinalizers([]string{finalizerName})
				c := mockclient.NewMockClient(ctrl)

				err := EnsureFinalizer(nil, c, finalizerName, secret)
				Expect(err).NotTo(HaveOccurred())
				Expect(secret.GetFinalizers()).To(Equal([]string{finalizerName}))
			})
		})

		Describe("#DeleteFinalizer", func() {
			It("should delete the finalizer if it was present", func() {
				secret := &corev1.Secret{}
				secret.SetFinalizers([]string{finalizerName})
				c := mockclient.NewMockClient(ctrl)
				c.EXPECT().Update(nil, secret)

				err := DeleteFinalizer(nil, c, finalizerName, secret)
				Expect(err).NotTo(HaveOccurred())
				Expect(secret.GetFinalizers()).To(BeEmpty())
			})

			It("should not delete the finalizer if it was not present", func() {
				secret := &corev1.Secret{}
				c := mockclient.NewMockClient(ctrl)

				err := DeleteFinalizer(nil, c, finalizerName, secret)
				Expect(err).NotTo(HaveOccurred())
				Expect(secret.GetFinalizers()).To(BeEmpty())
			})
		})

		Describe("#HasFinalizer", func() {
			It("should return true if the finalizer is present", func() {
				secret := &corev1.Secret{}
				secret.SetFinalizers([]string{finalizerName})

				hasFinalizer, err := HasFinalizer(secret, finalizerName)
				Expect(err).NotTo(HaveOccurred())
				Expect(hasFinalizer).To(BeTrue())
			})

			It("should return false if the finalizer was not present", func() {
				secret := &corev1.Secret{}

				hasFinalizer, err := HasFinalizer(secret, finalizerName)
				Expect(err).NotTo(HaveOccurred())
				Expect(hasFinalizer).To(BeFalse())
			})
		})
	})
})
