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

	"github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/config"
	"github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/openstack"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	"github.com/gardener/gardener-extensions/pkg/util"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

const (
	namespace = "test"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Openstack Controlplane Backup Webhook Suite")
}

var _ = Describe("Ensurer", func() {
	Describe("#EnsureETCDStatefulSet", func() {
		var (
			ctrl *gomock.Controller

			etcdBackup = &config.ETCDBackup{
				Schedule: util.StringPtr("0 */24 * * *"),
			}

			imageVector = imagevector.ImageVector{
				{
					Name:       openstack.ETCDBackupRestoreImageName,
					Repository: "test-repository",
					Tag:        util.StringPtr("test-tag"),
				},
			}

			cluster = &extensionscontroller.Cluster{
				Shoot: &gardenv1beta1.Shoot{
					Spec: gardenv1beta1.ShootSpec{
						Kubernetes: gardenv1beta1.Kubernetes{
							Version: "1.13.4",
						},
					},
					Status: gardenv1beta1.ShootStatus{
						TechnicalID: "shoot--test--sample",
						UID:         types.UID("test-uid"),
					},
				},
				Seed: &gardenv1beta1.Seed{
					Spec: gardenv1beta1.SeedSpec{
						Backup: &gardenv1beta1.BackupProfile{},
					},
				},
			}

			secretKey = client.ObjectKey{Namespace: namespace, Name: openstack.BackupSecretName}
			secret    = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: openstack.BackupSecretName, Namespace: namespace},
				Data:       map[string][]byte{"foo": []byte("bar")},
			}

			annotations = map[string]string{
				"checksum/secret-" + openstack.BackupSecretName: "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			}
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
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: gardencorev1alpha1.StatefulSetNameETCDMain},
				}
			)

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			// Create ensurer
			ensurer := NewEnsurer(etcdBackup, imageVector, logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureETCDStatefulSet method and check the result
			err = ensurer.EnsureETCDStatefulSet(context.TODO(), ss, cluster)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDMainStatefulSet(ss, annotations)
		})

		It("should modify existing elements of etcd-main statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: gardencorev1alpha1.StatefulSetNameETCDMain},
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

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			// Create ensurer
			ensurer := NewEnsurer(etcdBackup, imageVector, logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureETCDStatefulSet method and check the result
			err = ensurer.EnsureETCDStatefulSet(context.TODO(), ss, cluster)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDMainStatefulSet(ss, annotations)
		})

		It("should not configure backup to etcd-main statefulset if backup profile is missing", func() {
			ss := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: gardencorev1alpha1.StatefulSetNameETCDMain},
			}
			cluster.Seed.Spec.Backup = nil

			// Create mock client
			client := mockclient.NewMockClient(ctrl)

			// Create ensurer
			ensurer := NewEnsurer(etcdBackup, imageVector, logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureETCDStatefulSet method and check the result
			err = ensurer.EnsureETCDStatefulSet(context.TODO(), ss, cluster)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDMainStatefulSetWithoutBackup(ss, annotations)
		})

		It("should add or modify elements to etcd-events statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: gardencorev1alpha1.StatefulSetNameETCDEvents},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdBackup, imageVector, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), ss, cluster)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDEventsStatefulSet(ss)
		})

		It("should modify existing elements of etcd-events statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: gardencorev1alpha1.StatefulSetNameETCDEvents},
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
			ensurer := NewEnsurer(etcdBackup, imageVector, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), ss, cluster)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDEventsStatefulSet(ss)
		})
	})
})

func checkETCDMainStatefulSet(ss *appsv1.StatefulSet, annotations map[string]string) {
	var (
		env = []corev1.EnvVar{
			{
				Name: "STORAGE_CONTAINER",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key:                  openstack.BucketName,
						LocalObjectReference: corev1.LocalObjectReference{Name: openstack.BackupSecretName},
					},
				},
			},
			{
				Name: "OS_CLIENT_CONFIG_FILE",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key:                  openstack.CloudYAML,
						LocalObjectReference: corev1.LocalObjectReference{Name: openstack.BackupSecretName},
					},
				},
			},
			{
				Name: "OS_CLOUD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key:                  openstack.CloudYAMLKey,
						LocalObjectReference: corev1.LocalObjectReference{Name: openstack.BackupSecretName},
					},
				},
			},
		}
	)

	c := extensionswebhook.ContainerWithName(ss.Spec.Template.Spec.Containers, "backup-restore")
	Expect(c).To(Equal(controlplane.GetBackupRestoreContainer(gardencorev1alpha1.StatefulSetNameETCDMain, controlplane.EtcdMainVolumeClaimTemplateName, "0 */24 * * *", openstack.StorageProviderName, "shoot--test--sample--test-uid",
		"test-repository:test-tag", nil, env, nil)))
	Expect(ss.Spec.Template.Annotations).To(Equal(annotations))
}

func checkETCDMainStatefulSetWithoutBackup(ss *appsv1.StatefulSet, annotations map[string]string) {
	c := extensionswebhook.ContainerWithName(ss.Spec.Template.Spec.Containers, "backup-restore")
	Expect(c).To(Equal(controlplane.GetBackupRestoreContainer(gardencorev1alpha1.StatefulSetNameETCDMain, controlplane.EtcdMainVolumeClaimTemplateName, "0 */24 * * *", "", "",
		"test-repository:test-tag", nil, nil, nil)))
	Expect(ss.Spec.Template.Annotations).To(BeNil())
}

func checkETCDEventsStatefulSet(ss *appsv1.StatefulSet) {
	c := extensionswebhook.ContainerWithName(ss.Spec.Template.Spec.Containers, "backup-restore")
	Expect(c).To(Equal(controlplane.GetBackupRestoreContainer(gardencorev1alpha1.StatefulSetNameETCDEvents, gardencorev1alpha1.StatefulSetNameETCDEvents, "0 */24 * * *", "", "",
		"test-repository:test-tag", nil, nil, nil)))
}

func clientGet(result runtime.Object) interface{} {
	return func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
		switch obj.(type) {
		case *corev1.Secret:
			*obj.(*corev1.Secret) = *result.(*corev1.Secret)
		}
		return nil
	}
}
