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

package controlplanebackup

import (
	"context"
	"testing"

	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/packet"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/util"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/genericmutator"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	v1alpha1constants "github.com/gardener/gardener/pkg/apis/core/v1alpha1/constants"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace = "test"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Packet Controlplane Backup Webhook Suite")
}

var _ = Describe("Ensurer", func() {
	Describe("#EnsureETCDStatefulSet", func() {
		var (
			ctrl *gomock.Controller

			imageVector = imagevector.ImageVector{
				{
					Name:       packet.ETCDBackupRestoreImageName,
					Repository: "test-repository",
					Tag:        util.StringPtr("test-tag"),
				},
			}

			cluster = &extensionscontroller.Cluster{
				CoreShoot: &gardencorev1alpha1.Shoot{
					Spec: gardencorev1alpha1.ShootSpec{
						Kubernetes: gardencorev1alpha1.Kubernetes{
							Version: "1.13.4",
						},
					},
				},
			}

			dummyContext = genericmutator.NewInternalEnsurerContext(cluster)
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("should add or modify elements to etcd-main statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1alpha1constants.StatefulSetNameETCDMain},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(imageVector, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), dummyContext, ss)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDMainStatefulSet(ss, nil)
		})

		It("should modify existing elements of etcd-main statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1alpha1constants.StatefulSetNameETCDMain},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "backup-restore",
									},
								},
							},
						},
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(imageVector, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), dummyContext, ss)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDMainStatefulSet(ss, nil)
		})

		It("should not modify elements to same etcd-main statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1alpha1constants.StatefulSetNameETCDMain},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(imageVector, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), dummyContext, ss)
			Expect(err).To(Not(HaveOccurred()))
			oldSS := ss.DeepCopy()

			// Re-ensure on existing statefulset
			err = ensurer.EnsureETCDStatefulSet(context.TODO(), dummyContext, ss)

			Expect(err).To(Not(HaveOccurred()))
			Expect(ss).Should(Equal(oldSS))

			// Re-ensure on new statefulset request
			newSS := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1alpha1constants.StatefulSetNameETCDEvents},
			}
			err = ensurer.EnsureETCDStatefulSet(context.TODO(), dummyContext, newSS)

			Expect(err).To(Not(HaveOccurred()))
			Expect(ss).Should(Equal(oldSS))
		})

		It("should add or modify elements to etcd-events statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: v1alpha1constants.StatefulSetNameETCDEvents},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(imageVector, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), dummyContext, ss)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDEventsStatefulSet(ss)
		})

		It("should modify existing elements of etcd-events statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: v1alpha1constants.StatefulSetNameETCDEvents},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "backup-restore",
									},
								},
							},
						},
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(imageVector, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), dummyContext, ss)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDEventsStatefulSet(ss)
		})

		It("should not modify elements to same etcd-events statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1alpha1constants.StatefulSetNameETCDEvents},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(imageVector, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), dummyContext, ss)
			Expect(err).To(Not(HaveOccurred()))
			oldSS := ss.DeepCopy()

			// Re-ensure on existing statefulset
			err = ensurer.EnsureETCDStatefulSet(context.TODO(), dummyContext, ss)

			Expect(err).To(Not(HaveOccurred()))
			Expect(ss).Should(Equal(oldSS))

			// Re-ensure on new statefulset request
			newSS := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1alpha1constants.StatefulSetNameETCDEvents},
			}
			err = ensurer.EnsureETCDStatefulSet(context.TODO(), dummyContext, newSS)

			Expect(err).To(Not(HaveOccurred()))
			Expect(ss).Should(Equal(oldSS))
		})
	})
})

func checkETCDMainStatefulSet(ss *appsv1.StatefulSet, annotations map[string]string) {
	c := extensionswebhook.ContainerWithName(ss.Spec.Template.Spec.Containers, "backup-restore")
	Expect(c).To(Equal(controlplane.GetBackupRestoreContainer(v1alpha1constants.StatefulSetNameETCDMain, controlplane.EtcdMainVolumeClaimTemplateName, "", "", "",
		"test-repository:test-tag", nil, nil, nil)))
	Expect(ss.Spec.Template.Annotations).To(Equal(annotations))
}

func checkETCDEventsStatefulSet(ss *appsv1.StatefulSet) {
	c := extensionswebhook.ContainerWithName(ss.Spec.Template.Spec.Containers, "backup-restore")
	Expect(c).To(Equal(controlplane.GetBackupRestoreContainer(v1alpha1constants.StatefulSetNameETCDEvents, v1alpha1constants.StatefulSetNameETCDEvents, "", "", "",
		"test-repository:test-tag", nil, nil, nil)))
}
