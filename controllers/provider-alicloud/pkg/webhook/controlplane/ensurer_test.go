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

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/genericmutator"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/test"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"

	"github.com/coreos/go-systemd/unit"
	v1alpha1constants "github.com/gardener/gardener/pkg/apis/core/v1alpha1/constants"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Alicloud Controlplane Webhook Suite")
}

var _ = Describe("Ensurer", func() {
	var (
		ctrl       *gomock.Controller
		eContext13 = genericmutator.NewInternalEnsurerContext(
			&extensionscontroller.Cluster{
				Shoot: &gardencorev1alpha1.Shoot{
					Spec: gardencorev1alpha1.ShootSpec{
						Kubernetes: gardencorev1alpha1.Kubernetes{
							Version: "1.13.0",
						},
					},
				},
			},
		)
		eContext14 = genericmutator.NewInternalEnsurerContext(
			&extensionscontroller.Cluster{
				Shoot: &gardencorev1alpha1.Shoot{
					Spec: gardencorev1alpha1.ShootSpec{
						Kubernetes: gardencorev1alpha1.Kubernetes{
							Version: "1.14.0",
						},
					},
				},
			},
		)
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
				apidep = func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: v1alpha1constants.DeploymentNameKubeAPIServer},
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
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(logger)

			// Call EnsureKubeAPIServerDeployment method and check the result
			// 1.13
			dep := apidep()
			err := ensurer.EnsureKubeAPIServerDeployment(context.TODO(), eContext13, dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, []string{"CSINodeInfo=true", "CSIDriverRegistry=true"})
			// 1.14
			dep = apidep()
			err = ensurer.EnsureKubeAPIServerDeployment(context.TODO(), eContext14, dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, []string{"ExpandCSIVolumes=true", "ExpandInUsePersistentVolumes=true"})

		})

		It("should modify existing elements of kube-apiserver deployment", func() {
			var (
				apidep = func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: v1alpha1constants.DeploymentNameKubeAPIServer},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "kube-apiserver",
											Command: []string{
												"--enable-admission-plugins=Priority,PersistentVolumeLabel",
												"--disable-admission-plugins=",
												"--feature-gates=Foo=true",
											},
										},
									},
								},
							},
						},
					}
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(logger)

			// Call EnsureKubeAPIServerDeployment method and check the result
			// 1.13
			dep := apidep()
			err := ensurer.EnsureKubeAPIServerDeployment(context.TODO(), eContext13, dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, []string{"CSINodeInfo=true", "CSIDriverRegistry=true"})
			// 1.14
			dep = apidep()
			err = ensurer.EnsureKubeAPIServerDeployment(context.TODO(), eContext14, dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, []string{"ExpandCSIVolumes=true", "ExpandInUsePersistentVolumes=true"})
		})
	})

	Describe("#EnsureKubeControllerManagerDeployment", func() {
		It("should add missing elements to kube-controller-manager deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: v1alpha1constants.DeploymentNameKubeControllerManager},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-controller-manager",
									},
								},
							},
						},
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(logger)

			// Call EnsureKubeControllerManagerDeployment method and check the result
			err := ensurer.EnsureKubeControllerManagerDeployment(context.TODO(), eContext13, dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeControllerManagerDeployment(dep)
		})

		It("should modify existing elements of kube-controller-manager deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: v1alpha1constants.DeploymentNameKubeControllerManager},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-controller-manager",
										Command: []string{
											"--cloud-provider=?",
										},
									},
								},
							},
						},
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(logger)

			// Call EnsureKubeControllerManagerDeployment method and check the result
			err := ensurer.EnsureKubeControllerManagerDeployment(context.TODO(), eContext13, dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeControllerManagerDeployment(dep)
		})
	})

	Describe("#EnsureKubeletServiceUnitOptions", func() {
		It("should modify existing elements of kubelet.service unit options", func() {
			var (
				oldUnitOptions = []*unit.UnitOption{
					{
						Section: "Service",
						Name:    "ExecStart",
						Value: `/opt/bin/hyperkube kubelet \
    --config=/var/lib/kubelet/config/kubelet`,
					},
				}
				newUnitOptions = []*unit.UnitOption{
					{
						Section: "Service",
						Name:    "ExecStart",
						Value: `/opt/bin/hyperkube kubelet \
    --config=/var/lib/kubelet/config/kubelet \
    --provider-id=${PROVIDER_ID} \
    --cloud-provider=external \
    --enable-controller-attach-detach=true`,
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(logger)

			// Call EnsureKubeletServiceUnitOptions method and check the result
			opts, err := ensurer.EnsureKubeletServiceUnitOptions(context.TODO(), eContext13, oldUnitOptions)
			Expect(err).To(Not(HaveOccurred()))
			Expect(opts).To(Equal(newUnitOptions))
		})
	})

	Describe("#EnsureKubeletConfiguration", func() {
		It("should modify existing elements of kubelet configuration", func() {
			var (
				oldKubeletConfig13 = &kubeletconfigv1beta1.KubeletConfiguration{
					FeatureGates: map[string]bool{
						"Foo": true,
					},
				}
				newKubeletConfig13 = &kubeletconfigv1beta1.KubeletConfiguration{
					FeatureGates: map[string]bool{
						"Foo":               true,
						"CSINodeInfo":       true,
						"CSIDriverRegistry": true,
					},
				}
				oldKubeletConfig14 = &kubeletconfigv1beta1.KubeletConfiguration{
					FeatureGates: map[string]bool{
						"Foo": true,
					},
				}
				newKubeletConfig14 = &kubeletconfigv1beta1.KubeletConfiguration{
					FeatureGates: map[string]bool{
						"Foo":              true,
						"ExpandCSIVolumes": true,
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(logger)

			// Call EnsureKubeletConfiguration method and check the result
			// 13
			err := ensurer.EnsureKubeletConfiguration(context.TODO(), eContext13, oldKubeletConfig13)
			Expect(err).To(Not(HaveOccurred()))
			Expect(oldKubeletConfig13).To(Equal(newKubeletConfig13))
			// 14
			err = ensurer.EnsureKubeletConfiguration(context.TODO(), eContext14, oldKubeletConfig14)
			Expect(err).To(Not(HaveOccurred()))
			Expect(oldKubeletConfig14).To(Equal(newKubeletConfig14))
		})
	})
})

func checkKubeAPIServerDeployment(dep *appsv1.Deployment, featureGates []string) {
	// Check that the kube-apiserver container still exists and contains all needed command line args,
	// env vars, and volume mounts
	c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-apiserver")
	Expect(c).To(Not(BeNil()))
	Expect(c.Command).To(Not(test.ContainElementWithPrefixContaining("--enable-admission-plugins=", "PersistentVolumeLabel", ",")))
	Expect(c.Command).To(test.ContainElementWithPrefixContaining("--disable-admission-plugins=", "PersistentVolumeLabel", ","))
	for _, fg := range featureGates {
		Expect(c.Command).To(test.ContainElementWithPrefixContaining("--feature-gates=", fg, ","))
	}
}

func checkKubeControllerManagerDeployment(dep *appsv1.Deployment) {
	// Check that the kube-controller-manager container still exists and contains all needed command line args,
	// env vars, and volume mounts
	c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-controller-manager")
	Expect(c).To(Not(BeNil()))
	Expect(c.Command).To(ContainElement("--cloud-provider=external"))
}
