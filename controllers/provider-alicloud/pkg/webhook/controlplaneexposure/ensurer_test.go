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

package controlplaneexposure

import (
	"context"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/genericmutator"
	"testing"

	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/config"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	"github.com/gardener/gardener-extensions/pkg/util"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"

	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

const (
	namespace = "test"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Alicloud Controlplane Exposure Webhook Suite")
}

var _ = Describe("Ensurer", func() {
	var (
		etcdStorage = &config.ETCDStorage{
			ClassName: util.StringPtr("gardener.cloud-fast"),
			Capacity:  util.QuantityPtr(resource.MustParse("25Gi")),
		}

		ctrl *gomock.Controller

		svcKey = client.ObjectKey{Namespace: namespace, Name: common.KubeAPIServerDeploymentName}
		svc    = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: common.KubeAPIServerDeploymentName, Namespace: namespace},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{IP: "1.2.3.4"},
					},
				},
			},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#EnsureKubeAPIServerDeployment", func() {
		It("should add missing elements to kube-apiserver deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: common.KubeAPIServerDeploymentName, Namespace: namespace},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-apiserver",
									},
								},
							},
						},
					},
				}
			)

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), svcKey, &corev1.Service{}).DoAndReturn(clientGet(svc))

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err = ensurer.EnsureKubeAPIServerDeployment(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep)
		})

		It("should modify existing elements of kube-apiserver deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: common.KubeAPIServerDeploymentName, Namespace: namespace},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:    "kube-apiserver",
										Command: []string{"--advertise-address=?", "--external-hostname=?"},
									},
								},
							},
						},
					},
				}
			)

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), svcKey, &corev1.Service{}).DoAndReturn(clientGet(svc))

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err = ensurer.EnsureKubeAPIServerDeployment(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep)
		})
	})

	Describe("#EnsureKubeAPIServerNetworkPolicy", func() {
		It("should add missing elements to kube-apiserver network policy", func() {
			var (
				np = &networkingv1.NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{Name: genericmutator.KubeAPIServerNetworkPolicyName, Namespace: namespace},
					Spec: networkingv1.NetworkPolicySpec{
						Egress: []networkingv1.NetworkPolicyEgressRule{
							{
								To: []networkingv1.NetworkPolicyPeer{
									{
										IPBlock: &networkingv1.IPBlock{
											Except: []string{
												"172.16.0.0/20",
												"192.168.0.0/1",
												"10.0.0.0/24",
											},
										},
									},
								},
							},
						},
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			// Call EnsureKubeAPIServerNetworkPolicy method and check the result
			err := ensurer.EnsureKubeAPIServerNetworkPolicy(context.TODO(), np)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerNetworkPolicy(np)
		})

		It("should modify existing elements of kube-apiserver network policy", func() {
			var (
				np = &networkingv1.NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{Name: genericmutator.KubeAPIServerNetworkPolicyName, Namespace: namespace},
					Spec: networkingv1.NetworkPolicySpec{
						Egress: []networkingv1.NetworkPolicyEgressRule{
							{
								To: []networkingv1.NetworkPolicyPeer{
									{
										IPBlock: &networkingv1.IPBlock{
											Except: []string{
												"172.16.0.0/20",
												"192.168.0.0/1",
												"10.0.0.0/24",
												DefaultSeedCpCidr,
											},
										},
									},
								},
							},
						},
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			// Call EnsureKubeAPIServerNetworkPolicy method and check the result
			err := ensurer.EnsureKubeAPIServerNetworkPolicy(context.TODO(), np)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerNetworkPolicy(np)
		})
	})

	Describe("#EnsureETCDStatefulSet", func() {
		It("should add or modify elements to etcd-main statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: common.EtcdMainStatefulSetName},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), ss, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDMainStatefulSet(ss)
		})

		It("should modify existing elements of etcd-main statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: common.EtcdMainStatefulSetName},
					Spec: appsv1.StatefulSetSpec{
						VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
							{
								ObjectMeta: metav1.ObjectMeta{Name: "etcd-main"},
								Spec: corev1.PersistentVolumeClaimSpec{
									AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceStorage: resource.MustParse("10Gi"),
										},
									},
								},
							},
						},
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), ss, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDMainStatefulSet(ss)
		})

		It("should add or modify elements to etcd-events statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: common.EtcdEventsStatefulSetName},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), ss, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDEventsStatefulSet(ss)
		})

		It("should modify existing elements of etcd-events statefulset", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: common.EtcdEventsStatefulSetName},
					Spec: appsv1.StatefulSetSpec{
						VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
							{
								ObjectMeta: metav1.ObjectMeta{Name: "etcd-events"},
								Spec: corev1.PersistentVolumeClaimSpec{
									AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceStorage: resource.MustParse("20Gi"),
										},
									},
								},
							},
						},
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCDStatefulSet(context.TODO(), ss, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDEventsStatefulSet(ss)
		})
	})
})

func checkKubeAPIServerDeployment(dep *appsv1.Deployment) {
	// Check that the kube-apiserver container still exists and contains all needed command line args
	c := controlplane.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-apiserver")
	Expect(c).To(Not(BeNil()))
	Expect(c.Command).To(ContainElement("--advertise-address=1.2.3.4"))
	Expect(c.Command).To(ContainElement("--external-hostname=1.2.3.4"))
}

func checkETCDMainStatefulSet(ss *appsv1.StatefulSet) {
	pvc := controlplane.PVCWithName(ss.Spec.VolumeClaimTemplates, "etcd-main")
	Expect(pvc).To(Equal(controlplane.GetETCDVolumeClaimTemplate(common.EtcdMainStatefulSetName, util.StringPtr("gardener.cloud-fast"),
		util.QuantityPtr(resource.MustParse("25Gi")))))
}

func checkETCDEventsStatefulSet(ss *appsv1.StatefulSet) {
	pvc := controlplane.PVCWithName(ss.Spec.VolumeClaimTemplates, "etcd-events")
	Expect(pvc).To(Equal(controlplane.GetETCDVolumeClaimTemplate(common.EtcdEventsStatefulSetName, nil, nil)))
}

func checkKubeAPIServerNetworkPolicy(np *networkingv1.NetworkPolicy) {
	Expect(np.Spec.Egress).ToNot(BeEmpty())
	Expect(np.Spec.Egress[0].To).ToNot(BeEmpty())
	Expect(np.Spec.Egress[0].To[0].IPBlock.Except).ToNot(BeEmpty())
	Expect(np.Spec.Egress[0].To[0].IPBlock.Except).To(ContainElement(AlicloudCpCidr))
	Expect(np.Spec.Egress[0].To[0].IPBlock.Except).ToNot(ContainElement(DefaultSeedCpCidr))
}

func clientGet(result runtime.Object) interface{} {
	return func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
		switch obj.(type) {
		case *corev1.Service:
			*obj.(*corev1.Service) = *result.(*corev1.Service)
		}
		return nil
	}
}
