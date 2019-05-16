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

package genericmutator

import (
	"context"
	"testing"

	mockcontrolplane "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/webhook/controlplane"
	mockgenericmutator "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/webhook/controlplane/genericmutator"
	"github.com/gardener/gardener-extensions/pkg/util"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"

	"github.com/coreos/go-systemd/unit"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	oldServiceContent = "old kubelet.service content"
	newServiceContent = "new kubelet.service content"

	oldKubeletConfigData = "old kubelet config data"
	newKubeletConfigData = "new kubelet config data"
)

func TestControlplane(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controlplane Webhook Generic Mutator Suite")
}

var _ = Describe("Mutator", func() {
	var (
		ctrl   *gomock.Controller
		logger = log.Log.WithName("test")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#Mutate", func() {
		It("should invoke ensurer.EnsureKubeAPIServerService with a kube-apiserver service", func() {
			var (
				svc = &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: common.KubeAPIServerDeploymentName},
				}
			)

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureKubeAPIServerService(context.TODO(), svc).Return(nil)

			// Create mutator
			mutator := NewMutator(ensurer, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), svc)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should ignore other services than kube-apiserver", func() {
			var (
				svc = &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
				}
			)

			// Create mutator
			mutator := NewMutator(nil, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), svc)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should invoke ensurer.EnsureKubeAPIServerDeployment with a kube-apiserver deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: common.KubeAPIServerDeploymentName},
				}
			)

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureKubeAPIServerDeployment(context.TODO(), dep).Return(nil)

			// Create mutator
			mutator := NewMutator(ensurer, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should invoke ensurer.EnsureKubeControllerManagerDeployment with a kube-controller-manager deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: common.KubeControllerManagerDeploymentName},
				}
			)

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureKubeControllerManagerDeployment(context.TODO(), dep).Return(nil)

			// Create mutator
			mutator := NewMutator(ensurer, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should invoke ensurer.EnsureKubeSchedulerDeployment with a kube-scheduler deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: common.KubeSchedulerDeploymentName},
				}
			)

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureKubeSchedulerDeployment(context.TODO(), dep).Return(nil)

			// Create mutator
			mutator := NewMutator(ensurer, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should ignore other deployments than kube-apiserver, kube-controller-manager, and kube-scheduler", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
				}
			)

			// Create mutator
			mutator := NewMutator(nil, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should invoke appropriate ensurer methods with OperatingSystemConfig", func() {
			var (
				osc = &extensionsv1alpha1.OperatingSystemConfig{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
						Purpose: extensionsv1alpha1.OperatingSystemConfigPurposeReconcile,
						Units: []extensionsv1alpha1.Unit{
							{
								Name:    "kubelet.service",
								Content: util.StringPtr(oldServiceContent),
							},
						},
						Files: []extensionsv1alpha1.File{
							{
								Path: "/var/lib/kubelet/config/kubelet",
								Content: extensionsv1alpha1.FileContent{
									Inline: &extensionsv1alpha1.FileContentInline{
										Data: oldKubeletConfigData,
									},
								},
							},
						},
					},
				}

				oldUnitOptions = []*unit.UnitOption{
					{
						Section: "Service",
						Name:    "Foo",
						Value:   "bar",
					},
				}
				newUnitOptions = []*unit.UnitOption{
					{
						Section: "Service",
						Name:    "Foo",
						Value:   "baz",
					},
				}

				oldKubeletConfig = &kubeletconfigv1beta1.KubeletConfiguration{
					FeatureGates: map[string]bool{
						"Foo": true,
						"Bar": true,
					},
				}
				newKubeletConfig = &kubeletconfigv1beta1.KubeletConfiguration{
					FeatureGates: map[string]bool{
						"Foo": true,
					},
				}
			)

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureKubeletServiceUnitOptions(context.TODO(), oldUnitOptions).Return(newUnitOptions, nil)
			ensurer.EXPECT().EnsureKubeletConfiguration(context.TODO(), oldKubeletConfig).DoAndReturn(
				func(ctx context.Context, kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration) error {
					*kubeletConfig = *newKubeletConfig
					return nil
				},
			)

			// Create mock UnitSerializer
			us := mockcontrolplane.NewMockUnitSerializer(ctrl)
			us.EXPECT().Deserialize(oldServiceContent).Return(oldUnitOptions, nil)
			us.EXPECT().Serialize(newUnitOptions).Return(newServiceContent, nil)

			// Create mock KubeletConfigCodec
			kcc := mockcontrolplane.NewMockKubeletConfigCodec(ctrl)
			kcc.EXPECT().Decode(&extensionsv1alpha1.FileContentInline{Data: oldKubeletConfigData}).Return(oldKubeletConfig, nil)
			kcc.EXPECT().Encode(newKubeletConfig, "").Return(&extensionsv1alpha1.FileContentInline{Data: newKubeletConfigData}, nil)

			// Create mutator
			mutator := NewMutator(ensurer, us, kcc, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), osc)
			Expect(err).To(Not(HaveOccurred()))
			checkOperatingSystemConfig(osc)
		})
	})
})

func checkOperatingSystemConfig(osc *extensionsv1alpha1.OperatingSystemConfig) {
	u := controlplane.UnitWithName(osc.Spec.Units, "kubelet.service")
	Expect(u).To(Not(BeNil()))
	Expect(u.Content).To(Equal(util.StringPtr(newServiceContent)))
	f := controlplane.FileWithPath(osc.Spec.Files, "/var/lib/kubelet/config/kubelet")
	Expect(f).To(Not(BeNil()))
	Expect(f.Content.Inline).To(Equal(&extensionsv1alpha1.FileContentInline{Data: newKubeletConfigData}))
}
