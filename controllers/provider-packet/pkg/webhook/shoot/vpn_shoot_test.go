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

package shoot

import (
	"context"

	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Mutator", func() {
	Describe("#mutateVPNShootDeployment", func() {
		It("should correctly inject the init container", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vpn-shoot",
					Namespace: metav1.NamespaceSystem,
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "vpn-shoot",
								},
							},
						},
					},
				},
			}

			mutator := &mutator{}
			err := mutator.mutateVPNShootDeployment(context.TODO(), deployment)

			Expect(err).To(Not(HaveOccurred()))
			checkVPNShootDeployment(deployment)
		})
	})
})

func checkVPNShootDeployment(deployment *appsv1.Deployment) {
	// Check init container
	ic := extensionswebhook.ContainerWithName(deployment.Spec.Template.Spec.InitContainers, metabotInitContainerName)
	Expect(ic).To(Not(BeNil()))
	Expect(ic.Name).To(Equal(metabotInitContainerName))
	Expect(ic.Image).ToNot(BeNil())
	Expect(extensionswebhook.StringWithPrefixIndex(ic.Args, "ip")).NotTo(Equal(-1))
	Expect(extensionswebhook.StringWithPrefixIndex(ic.Args, "4")).NotTo(Equal(-1))
	Expect(extensionswebhook.StringWithPrefixIndex(ic.Args, "private")).NotTo(Equal(-1))
	Expect(extensionswebhook.StringWithPrefixIndex(ic.Args, "parent")).NotTo(Equal(-1))
	Expect(extensionswebhook.StringWithPrefixIndex(ic.Args, "network")).NotTo(Equal(-1))
	Expect(ic.VolumeMounts).To(ContainElement(corev1.VolumeMount{
		Name:      metabotVolumeName,
		MountPath: metabotVolumeMountPath,
	}))

	// Check that the vpn-shoot container still exists and contains all needed volume mounts
	c := extensionswebhook.ContainerWithName(deployment.Spec.Template.Spec.Containers, "vpn-shoot")
	Expect(c).To(Not(BeNil()))
	Expect(c.VolumeMounts).To(ContainElement(corev1.VolumeMount{
		Name:      metabotVolumeName,
		MountPath: metabotVolumeMountPath,
	}))

	// Check that the pod spec contains all needed volumes
	Expect(deployment.Spec.Template.Spec.Volumes).To(ContainElement(corev1.Volume{
		Name: metabotVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}))
}
