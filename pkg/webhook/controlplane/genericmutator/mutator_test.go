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

package genericmutator_test

import (
	"context"
	"encoding/json"
	"testing"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockcontrolplane "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/webhook/controlplane"
	mockgenericmutator "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/webhook/controlplane/genericmutator"
	"github.com/gardener/gardener-extensions/pkg/util"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/genericmutator"

	"github.com/coreos/go-systemd/unit"
	gardencorevalpha1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

const (
	oldServiceContent = "old kubelet.service content"
	newServiceContent = "new kubelet.service content"

	oldKubeletConfigData = "old kubelet config data"
	newKubeletConfigData = "new kubelet config data"

	oldKubernetesGeneralConfigData = "# Increase the tcp-time-wait buckets pool size to prevent simple DOS attacks\nnet.ipv4.tcp_tw_reuse = 1"
	newKubernetesGeneralConfigData = "# Increase the tcp-time-wait buckets pool size to prevent simple DOS attacks\nnet.ipv4.tcp_tw_reuse = 1\n# Provider specific settings"

	encoding                 = "b64"
	cloudproviderconf        = "[Global]\nauth-url: whatever-url/keystone"
	cloudproviderconfEncoded = "W0dsb2JhbF1cbmF1dGgtdXJsOiBodHRwczovL2NsdXN0ZXIuZXUtZGUtMjAwLmNsb3VkLnNhcDo1MDAwL3Yz"
)

const (
	namespace = "test"
)

func TestControlplane(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controlplane Webhook Generic Mutator Suite")
}

var _ = Describe("Mutator", func() {
	var (
		ctrl   *gomock.Controller
		logger = log.Log.WithName("test")

		clusterKey = client.ObjectKey{Name: namespace}
		cluster    = &extensionscontroller.Cluster{
			CloudProfile: &gardencorevalpha1.CloudProfile{
				TypeMeta: metav1.TypeMeta{
					APIVersion: gardencorevalpha1.SchemeGroupVersion.String(),
					Kind:       "CloudProfile",
				},
			},
			Seed: &gardencorevalpha1.Seed{
				TypeMeta: metav1.TypeMeta{
					APIVersion: gardencorevalpha1.SchemeGroupVersion.String(),
					Kind:       "Seed",
				},
			},
			Shoot: &gardencorevalpha1.Shoot{
				TypeMeta: metav1.TypeMeta{
					APIVersion: gardencorevalpha1.SchemeGroupVersion.String(),
					Kind:       "Shoot",
				},
				Spec: gardencorevalpha1.ShootSpec{
					Kubernetes: gardencorevalpha1.Kubernetes{
						Version: "1.13.4",
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

	Describe("#Mutate", func() {
		It("should invoke ensurer.EnsureKubeAPIServerService with a kube-apiserver service", func() {
			var (
				svc = &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.DeploymentNameKubeAPIServer},
				}
			)

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureKubeAPIServerService(context.TODO(), gomock.Any(), svc).Return(nil)

			// Create mutator
			mutator := genericmutator.NewMutator(ensurer, nil, nil, nil, logger)

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
			mutator := genericmutator.NewMutator(nil, nil, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), svc)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should invoke ensurer.EnsureKubeAPIServerDeployment with a kube-apiserver deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.DeploymentNameKubeAPIServer},
				}
			)

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureKubeAPIServerDeployment(context.TODO(), gomock.Any(), dep).Return(nil)

			// Create mutator
			mutator := genericmutator.NewMutator(ensurer, nil, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should invoke ensurer.EnsureKubeControllerManagerDeployment with a kube-controller-manager deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.DeploymentNameKubeControllerManager},
				}
			)

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureKubeControllerManagerDeployment(context.TODO(), gomock.Any(), dep).Return(nil)

			// Create mutator
			mutator := genericmutator.NewMutator(ensurer, nil, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should invoke ensurer.EnsureKubeSchedulerDeployment with a kube-scheduler deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.DeploymentNameKubeScheduler},
				}
			)

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureKubeSchedulerDeployment(context.TODO(), gomock.Any(), dep).Return(nil)

			// Create mutator
			mutator := genericmutator.NewMutator(ensurer, nil, nil, nil, logger)

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
			mutator := genericmutator.NewMutator(nil, nil, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should invoke ensurer.EnsureETCDStatefulSet with a etcd-main stateful set", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.ETCDMain, Namespace: namespace},
				}
			)

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), clusterKey, &extensionsv1alpha1.Cluster{}).DoAndReturn(clientGet(clusterObject(cluster)))

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureETCDStatefulSet(context.TODO(), gomock.Any(), ss).Return(nil).Do(func(ctx context.Context, ectx genericmutator.EnsurerContext, ss *appsv1.StatefulSet) {
				_, err := ectx.GetCluster(ctx)
				if err != nil {
					logger.Error(err, "failed to get cluster object")
				}
			})

			// Create mutator
			mutator := genericmutator.NewMutator(ensurer, nil, nil, nil, logger)
			err := mutator.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call Mutate method and check the result
			err = mutator.Mutate(context.TODO(), ss)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should invoke ensurer.EnsureETCDStatefulSet with a etcd-events stateful set", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.ETCDEvents, Namespace: namespace},
				}
			)

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), clusterKey, &extensionsv1alpha1.Cluster{}).DoAndReturn(clientGet(clusterObject(cluster)))

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureETCDStatefulSet(context.TODO(), gomock.Any(), ss).Return(nil).Do(func(ctx context.Context, ectx genericmutator.EnsurerContext, ss *appsv1.StatefulSet) {
				_, err := ectx.GetCluster(ctx)
				if err != nil {
					logger.Error(err, "failed to get cluster object")
				}
			})

			// Create mutator
			mutator := genericmutator.NewMutator(ensurer, nil, nil, nil, logger)
			err := mutator.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call Mutate method and check the result
			err = mutator.Mutate(context.TODO(), ss)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should ignore other stateful sets than etcd-main and etcd-events", func() {
			var (
				ss = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
				}
			)

			// Create mutator
			mutator := genericmutator.NewMutator(nil, nil, nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), ss)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should invoke appropriate ensurer methods with OperatingSystemConfig", func() {
			var (
				osc = &extensionsv1alpha1.OperatingSystemConfig{
					ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
					Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
						Purpose: extensionsv1alpha1.OperatingSystemConfigPurposeReconcile,
						Units: []extensionsv1alpha1.Unit{
							{
								Name:    v1beta1constants.OperatingSystemConfigUnitNameKubeletService,
								Content: util.StringPtr(oldServiceContent),
							},
						},
						Files: []extensionsv1alpha1.File{
							{
								Path: v1beta1constants.OperatingSystemConfigFilePathKubeletConfig,
								Content: extensionsv1alpha1.FileContent{
									Inline: &extensionsv1alpha1.FileContentInline{
										Data: oldKubeletConfigData,
									},
								},
							},
							{
								Path: v1beta1constants.OperatingSystemConfigFilePathKernelSettings,
								Content: extensionsv1alpha1.FileContent{
									Inline: &extensionsv1alpha1.FileContentInline{
										Data: oldKubernetesGeneralConfigData,
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
				additionalUnit = extensionsv1alpha1.Unit{Name: "custom-mtu.service"}
				additionalFile = extensionsv1alpha1.File{Path: "/test/path"}
			)

			// Create mock ensurer
			ensurer := mockgenericmutator.NewMockEnsurer(ctrl)
			ensurer.EXPECT().EnsureKubeletServiceUnitOptions(context.TODO(), gomock.Any(), oldUnitOptions).Return(newUnitOptions, nil)
			ensurer.EXPECT().EnsureKubeletConfiguration(context.TODO(), gomock.Any(), oldKubeletConfig).DoAndReturn(
				func(ctx context.Context, ectx genericmutator.EnsurerContext, kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration) error {
					*kubeletConfig = *newKubeletConfig
					return nil
				},
			)
			ensurer.EXPECT().EnsureKubernetesGeneralConfiguration(context.TODO(), gomock.Any(), util.StringPtr(oldKubernetesGeneralConfigData)).DoAndReturn(
				func(ctx context.Context, ectx genericmutator.EnsurerContext, data *string) error {
					*data = newKubernetesGeneralConfigData
					return nil
				},
			)
			ensurer.EXPECT().EnsureAdditionalUnits(context.TODO(), gomock.Any(), &osc.Spec.Units).DoAndReturn(
				func(ctx context.Context, ectx genericmutator.EnsurerContext, oscUnits *[]extensionsv1alpha1.Unit) error {
					*oscUnits = append(*oscUnits, additionalUnit)
					return nil
				})
			ensurer.EXPECT().EnsureAdditionalFiles(context.TODO(), gomock.Any(), &osc.Spec.Files).DoAndReturn(
				func(ctx context.Context, ectx genericmutator.EnsurerContext, oscFiles *[]extensionsv1alpha1.File) error {
					*oscFiles = append(*oscFiles, additionalFile)
					return nil
				})

			ensurer.EXPECT().ShouldProvisionKubeletCloudProviderConfig().Return(true)
			ensurer.EXPECT().EnsureKubeletCloudProviderConfig(context.TODO(), gomock.Any(), util.StringPtr(""), osc.Namespace).DoAndReturn(
				func(ctx context.Context, ectx genericmutator.EnsurerContext, data *string, _ string) error {
					*data = cloudproviderconf
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

			// Create mock FileContentInlineCodec
			fcic := mockcontrolplane.NewMockFileContentInlineCodec(ctrl)
			fcic.EXPECT().Decode(&extensionsv1alpha1.FileContentInline{Data: oldKubernetesGeneralConfigData}).Return([]byte(oldKubernetesGeneralConfigData), nil)
			fcic.EXPECT().Encode([]byte(newKubernetesGeneralConfigData), "").Return(&extensionsv1alpha1.FileContentInline{Data: newKubernetesGeneralConfigData}, nil)
			fcic.EXPECT().Encode([]byte(cloudproviderconf), encoding).Return(&extensionsv1alpha1.FileContentInline{Data: cloudproviderconfEncoded, Encoding: encoding}, nil)

			// Create mutator
			mutator := genericmutator.NewMutator(ensurer, us, kcc, fcic, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), osc)
			Expect(err).To(Not(HaveOccurred()))
			checkOperatingSystemConfig(osc)
		})
	})
})

func checkOperatingSystemConfig(osc *extensionsv1alpha1.OperatingSystemConfig) {
	kubeletUnit := extensionswebhook.UnitWithName(osc.Spec.Units, v1beta1constants.OperatingSystemConfigUnitNameKubeletService)
	Expect(kubeletUnit).To(Not(BeNil()))
	Expect(kubeletUnit.Content).To(Equal(util.StringPtr(newServiceContent)))

	customMTU := extensionswebhook.UnitWithName(osc.Spec.Units, "custom-mtu.service")
	Expect(customMTU).To(Not(BeNil()))

	customFile := extensionswebhook.FileWithPath(osc.Spec.Files, "/test/path")
	Expect(customFile).To(Not(BeNil()))

	kubeletFile := extensionswebhook.FileWithPath(osc.Spec.Files, v1beta1constants.OperatingSystemConfigFilePathKubeletConfig)
	Expect(kubeletFile).To(Not(BeNil()))
	Expect(kubeletFile.Content.Inline).To(Equal(&extensionsv1alpha1.FileContentInline{Data: newKubeletConfigData}))

	general := extensionswebhook.FileWithPath(osc.Spec.Files, v1beta1constants.OperatingSystemConfigFilePathKernelSettings)
	Expect(general).To(Not(BeNil()))
	Expect(general.Content.Inline).To(Equal(&extensionsv1alpha1.FileContentInline{Data: newKubernetesGeneralConfigData}))

	cloudProvider := extensionswebhook.FileWithPath(osc.Spec.Files, genericmutator.CloudProviderConfigPath)
	Expect(cloudProvider).To(Not(BeNil()))
	Expect(cloudProvider.Path).To(Equal(genericmutator.CloudProviderConfigPath))
	Expect(cloudProvider.Permissions).To(Equal(util.Int32Ptr(0644)))
	Expect(cloudProvider.Content.Inline).To(Equal(&extensionsv1alpha1.FileContentInline{Data: cloudproviderconfEncoded, Encoding: encoding}))
}

func clientGet(result runtime.Object) interface{} {
	return func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
		switch obj.(type) {
		case *extensionsv1alpha1.Cluster:
			*obj.(*extensionsv1alpha1.Cluster) = *result.(*extensionsv1alpha1.Cluster)
		}
		return nil
	}
}

func clusterObject(cluster *extensionscontroller.Cluster) *extensionsv1alpha1.Cluster {
	return &extensionsv1alpha1.Cluster{
		Spec: extensionsv1alpha1.ClusterSpec{
			CloudProfile: runtime.RawExtension{
				Raw: encode(cluster.CloudProfile),
			},
			Seed: runtime.RawExtension{
				Raw: encode(cluster.Seed),
			},
			Shoot: runtime.RawExtension{
				Raw: encode(cluster.Shoot),
			},
		},
	}
}

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}
