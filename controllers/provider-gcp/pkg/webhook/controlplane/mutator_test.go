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

	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal"
	mockcontrolplane "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/webhook/controlplane"
	"github.com/gardener/gardener-extensions/pkg/util"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/test"

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
)

const (
	oldServiceContent = "old kubelet.service content"
	newServiceContent = "new kubelet.service content"

	oldKubeletConfigData = "old kubelet config data"
	newKubeletConfigData = "new kubelet config data"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GCP Controlplane Webhook Suite")
}

var _ = Describe("Mutator", func() {
	var (
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#Mutate", func() {
		It("should add missing elements to kube-apiserver deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: common.KubeAPIServerDeploymentName},
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

			// Create mutator
			mutator := NewMutator(nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep)
		})

		It("should modify existing elements of kube-apiserver deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: common.KubeAPIServerDeploymentName},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-apiserver",
										Command: []string{
											"--cloud-provider=?",
											"--cloud-config=?",
											"--enable-admission-plugins=Priority,NamespaceLifecycle",
											"--disable-admission-plugins=PersistentVolumeLabel",
										},
										Env: []corev1.EnvVar{
											{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "?"},
										},
										VolumeMounts: []corev1.VolumeMount{
											{Name: internal.CloudProviderConfigName, MountPath: "?"},
											// TODO Use constant from github.com/gardener/gardener/pkg/apis/core/v1alpha1 when available
											// See https://github.com/gardener/gardener/pull/930
											{Name: common.CloudProviderSecretName, MountPath: "?"},
										},
									},
								},
								Volumes: []corev1.Volume{
									{Name: internal.CloudProviderConfigName},
									{Name: common.CloudProviderSecretName},
								},
							},
						},
					},
				}
			)

			// Create mutator
			mutator := NewMutator(nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep)
		})

		It("should add missing elements to kube-controller-manager deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: common.KubeControllerManagerDeploymentName},
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

			// Create mutator
			mutator := NewMutator(nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeControllerManagerDeployment(dep)
		})

		It("should modify existing elements of kube-controller-manager deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: common.KubeControllerManagerDeploymentName},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-controller-manager",
										Command: []string{
											"--cloud-provider=?",
											"--cloud-config=?",
											"--external-cloud-volume-plugin=?",
										},
										Env: []corev1.EnvVar{
											{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "?"},
										},
										VolumeMounts: []corev1.VolumeMount{
											{Name: internal.CloudProviderConfigName, MountPath: "?"},
											{Name: common.CloudProviderSecretName, MountPath: "?"},
										},
									},
								},
								Volumes: []corev1.Volume{
									{Name: internal.CloudProviderConfigName},
									{Name: common.CloudProviderSecretName},
								},
							},
						},
					},
				}
			)

			// Create mutator
			mutator := NewMutator(nil, nil, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeControllerManagerDeployment(dep)
		})

		It("should modify existing elements of OperatingSystemConfig", func() {
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
    --cloud-provider=gce`,
					},
					{
						Section: "Service",
						Name:    "ExecStartPre",
						Value:   `/bin/sh -c 'hostnamectl set-hostname $(echo $HOSTNAME | cut -d '.' -f 1)'`,
					},
				}

				oldKubeletConfig = &kubeletconfigv1beta1.KubeletConfiguration{
					FeatureGates: map[string]bool{
						"Foo":                      true,
						"VolumeSnapshotDataSource": true,
						"CSINodeInfo":              true,
					},
				}
				newKubeletConfig = &kubeletconfigv1beta1.KubeletConfiguration{
					FeatureGates: map[string]bool{
						"Foo": true,
					},
				}
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
			mutator := NewMutator(us, kcc, logger)

			// Call Mutate method and check the result
			err := mutator.Mutate(context.TODO(), osc)
			Expect(err).To(Not(HaveOccurred()))
			checkOperatingSystemConfig(osc)
		})
	})
})

func checkKubeAPIServerDeployment(dep *appsv1.Deployment) {
	// Check that the kube-apiserver container still exists and contains all needed command line args,
	// env vars, and volume mounts
	c := controlplane.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-apiserver")
	Expect(c).To(Not(BeNil()))
	Expect(c.Command).To(ContainElement("--cloud-provider=gce"))
	Expect(c.Command).To(ContainElement("--cloud-config=/etc/kubernetes/cloudprovider/cloudprovider.conf"))
	Expect(c.Command).To(test.ContainElementWithPrefixContaining("--enable-admission-plugins=", "PersistentVolumeLabel", ","))
	Expect(c.Command).To(Not(test.ContainElementWithPrefixContaining("--disable-admission-plugins=", "PersistentVolumeLabel", ",")))
	Expect(c.Env).To(ContainElement(credentialsEnvVar))
	Expect(c.VolumeMounts).To(ContainElement(cloudProviderConfigVolumeMount))
	Expect(c.VolumeMounts).To(ContainElement(cloudProviderSecretVolumeMount))

	// Check that the Pod spec contains all needed volumes
	Expect(dep.Spec.Template.Spec.Volumes).To(ContainElement(cloudProviderConfigVolume))
	Expect(dep.Spec.Template.Spec.Volumes).To(ContainElement(cloudProviderSecretVolume))
}

func checkKubeControllerManagerDeployment(dep *appsv1.Deployment) {
	// Check that the kube-controller-manager container still exists and contains all needed command line args,
	// env vars, and volume mounts
	c := controlplane.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-controller-manager")
	Expect(c).To(Not(BeNil()))
	Expect(c.Command).To(ContainElement("--cloud-provider=external"))
	Expect(c.Command).To(ContainElement("--cloud-config=/etc/kubernetes/cloudprovider/cloudprovider.conf"))
	Expect(c.Command).To(ContainElement("--external-cloud-volume-plugin=gce"))
	Expect(c.Env).To(ContainElement(credentialsEnvVar))
	Expect(c.VolumeMounts).To(ContainElement(cloudProviderConfigVolumeMount))
	Expect(c.VolumeMounts).To(ContainElement(cloudProviderSecretVolumeMount))

	// Check that the Pod spec contains all needed volumes
	Expect(dep.Spec.Template.Spec.Volumes).To(ContainElement(cloudProviderConfigVolume))
	Expect(dep.Spec.Template.Spec.Volumes).To(ContainElement(cloudProviderSecretVolume))
}

func checkOperatingSystemConfig(osc *extensionsv1alpha1.OperatingSystemConfig) {
	u := controlplane.UnitWithName(osc.Spec.Units, "kubelet.service")
	Expect(u).To(Not(BeNil()))
	Expect(u.Content).To(Equal(util.StringPtr(newServiceContent)))
	f := controlplane.FileWithPath(osc.Spec.Files, "/var/lib/kubelet/config/kubelet")
	Expect(f).To(Not(BeNil()))
	Expect(f.Content.Inline).To(Equal(&extensionsv1alpha1.FileContentInline{Data: newKubeletConfigData}))
}
