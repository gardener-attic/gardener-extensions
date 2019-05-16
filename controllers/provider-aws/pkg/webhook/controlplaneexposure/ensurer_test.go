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
	"testing"

	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"

	"github.com/gardener/gardener/pkg/operation/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS Controlplane Exposure Webhook Suite")
}

var _ = Describe("Mutator", func() {

	Describe("#EnsureKubeAPIServerService", func() {
		It("should add annotations to kube-apiserver service", func() {
			var (
				svc = &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "kube-apiserver"},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(logger)

			// Call EnsureKubeAPIServerService method and check the result
			err := ensurer.EnsureKubeAPIServerService(context.TODO(), svc)
			Expect(err).To(Not(HaveOccurred()))
			Expect(svc.Annotations).To(HaveKeyWithValue("service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout", "3600"))
			Expect(svc.Annotations).To(HaveKeyWithValue("service.beta.kubernetes.io/aws-load-balancer-backend-protocol", "ssl"))
			Expect(svc.Annotations).To(HaveKeyWithValue("service.beta.kubernetes.io/aws-load-balancer-ssl-ports", "443"))
			Expect(svc.Annotations).To(HaveKeyWithValue("service.beta.kubernetes.io/aws-load-balancer-healthcheck-timeout", "5"))
			Expect(svc.Annotations).To(HaveKeyWithValue("service.beta.kubernetes.io/aws-load-balancer-healthcheck-interval", "30"))
			Expect(svc.Annotations).To(HaveKeyWithValue("service.beta.kubernetes.io/aws-load-balancer-healthcheck-healthy-threshold", "2"))
			Expect(svc.Annotations).To(HaveKeyWithValue("service.beta.kubernetes.io/aws-load-balancer-healthcheck-unhealthy-threshold", "2"))
			Expect(svc.Annotations).To(HaveKeyWithValue("service.beta.kubernetes.io/aws-load-balancer-ssl-negotiation-policy", "ELBSecurityPolicy-TLS-1-2-2017-01"))
		})
	})

	Describe("#EnsureKubeAPIServerDeployment", func() {
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

			// Create ensurer
			ensurer := NewEnsurer(logger)

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureKubeAPIServerDeployment(context.TODO(), dep)
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
										Name:    "kube-apiserver",
										Command: []string{"--endpoint-reconciler-type=?"},
									},
								},
							},
						},
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(logger)

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureKubeAPIServerDeployment(context.TODO(), dep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep)
		})
	})
})

func checkKubeAPIServerDeployment(dep *appsv1.Deployment) {
	// Check that the kube-apiserver container still exists and contains all needed command line args
	c := controlplane.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-apiserver")
	Expect(c).To(Not(BeNil()))
	Expect(c.Command).To(ContainElement("--endpoint-reconciler-type=none"))
}
