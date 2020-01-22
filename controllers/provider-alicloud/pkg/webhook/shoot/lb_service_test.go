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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Mutator", func() {
	var (
		mutator = NewMutator()
		vpnSvc  = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vpn-shoot",
				Namespace: metav1.NamespaceSystem},
			Spec: corev1.ServiceSpec{ExternalTrafficPolicy: "Cluster"}}
		nginxIngressSvc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "addons-nginx-ingress-controller",
				Namespace: metav1.NamespaceSystem},
			Spec: corev1.ServiceSpec{ExternalTrafficPolicy: "Cluster"}}
		otherSvc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other",
				Namespace: metav1.NamespaceSystem},
			Spec: corev1.ServiceSpec{ExternalTrafficPolicy: "Cluster"}}
	)
	Describe("#MutateLBService", func() {
		It("should set ExternalTrafficPolicy to Local for VPN service", func() {
			err := mutator.Mutate(context.TODO(), vpnSvc)
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(vpnSvc.Spec.ExternalTrafficPolicy)).Should(Equal("Local"))
		})
		It("should set ExternalTrafficPolicy to Nginx Ingress for VPN service", func() {
			err := mutator.Mutate(context.TODO(), nginxIngressSvc)
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(nginxIngressSvc.Spec.ExternalTrafficPolicy)).Should(Equal("Local"))
		})
		It("should not set ExternalTrafficPolicy to Nginx Ingress for Other service", func() {
			err := mutator.Mutate(context.TODO(), otherSvc)
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(otherSvc.Spec.ExternalTrafficPolicy)).Should(Equal("Cluster"))
		})
	})
})
