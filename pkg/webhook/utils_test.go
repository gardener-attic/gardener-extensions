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

package webhook

import (
	"testing"

	mockmeta "github.com/gardener/gardener-extensions/pkg/mock/apimachinery/api/meta"
	mockmanager "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/manager"
	mockadmission "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/webhook/admission"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/types"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook Suite")
}

var _ = Describe("Options", func() {
	const (
		provider = "aws"
	)

	var (
		ctrl    *gomock.Controller
		scheme  *runtime.Scheme
		mapper  *mockmeta.MockRESTMapper
		mgr     *mockmanager.MockManager
		handler *mockadmission.MockHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		// Build scheme
		scheme = runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)
		_ = appsv1.AddToScheme(scheme)

		// Create mock handler
		handler = mockadmission.NewMockHandler(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#NewWebhook", func() {
		It("should create the correct Shoot webhook for deployments", func() {
			// Create mock RESTMapper
			mapper = mockmeta.NewMockRESTMapper(ctrl)
			mapper.EXPECT().RESTMapping(schema.GroupKind{Group: "apps", Kind: "Deployment"}, "v1").Return(&meta.RESTMapping{
				Resource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			}, nil)

			// Create mock manager
			mgr = mockmanager.NewMockManager(ctrl)
			mgr.EXPECT().GetScheme().Return(scheme)
			mgr.EXPECT().GetRESTMapper().Return(mapper)

			webhook, err := NewWebhook(mgr, ShootKind, provider, "controlplane", []runtime.Object{&appsv1.Deployment{}}, handler)
			Expect(err).NotTo(HaveOccurred())
			Expect(webhook).To(Equal(&admission.Webhook{
				Name: "controlplane.aws.extensions.gardener.cloud",
				Type: types.WebhookTypeMutating,
				Path: "/controlplane",
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					ruleWithOperations("apps", "v1", "deployments"),
				},
				FailurePolicy: failurePolicyTypePtr(admissionregistrationv1beta1.Fail),
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{Key: ShootProviderLabel, Operator: metav1.LabelSelectorOpIn, Values: []string{provider}},
					},
				},
				Handlers: []admission.Handler{handler},
			}))
		})

		It("should create the correct Seed webhook for services and deployments", func() {
			// Create mock RESTMapper
			mapper = mockmeta.NewMockRESTMapper(ctrl)
			mapper.EXPECT().RESTMapping(schema.GroupKind{Group: "", Kind: "Service"}, "v1").Return(&meta.RESTMapping{
				Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"},
			}, nil)
			mapper.EXPECT().RESTMapping(schema.GroupKind{Group: "apps", Kind: "Deployment"}, "v1").Return(&meta.RESTMapping{
				Resource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			}, nil)

			// Create mock manager
			mgr = mockmanager.NewMockManager(ctrl)
			mgr.EXPECT().GetScheme().Return(scheme).Times(2)
			mgr.EXPECT().GetRESTMapper().Return(mapper).Times(2)

			webhook, err := NewWebhook(mgr, SeedKind, provider, "controlplaneexposure", []runtime.Object{&corev1.Service{}, &appsv1.Deployment{}}, handler)
			Expect(err).NotTo(HaveOccurred())
			Expect(webhook).To(Equal(&admission.Webhook{
				Name: "controlplaneexposure.aws.extensions.gardener.cloud",
				Type: types.WebhookTypeMutating,
				Path: "/controlplaneexposure",
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					ruleWithOperations("", "v1", "services"),
					ruleWithOperations("apps", "v1", "deployments"),
				},
				FailurePolicy: failurePolicyTypePtr(admissionregistrationv1beta1.Fail),
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{Key: SeedProviderLabel, Operator: metav1.LabelSelectorOpIn, Values: []string{provider}},
					},
				},
				Handlers: []admission.Handler{handler},
			}))
		})
	})
})

func ruleWithOperations(apiGroup, apiVersion, resource string) admissionregistrationv1beta1.RuleWithOperations {
	return admissionregistrationv1beta1.RuleWithOperations{
		Operations: []admissionregistrationv1beta1.OperationType{
			admissionregistrationv1beta1.Create,
			admissionregistrationv1beta1.Update,
		},
		Rule: admissionregistrationv1beta1.Rule{
			APIGroups:   []string{apiGroup},
			APIVersions: []string{apiVersion},
			Resources:   []string{resource},
		},
	}
}

func failurePolicyTypePtr(fp admissionregistrationv1beta1.FailurePolicyType) *admissionregistrationv1beta1.FailurePolicyType {
	return &fp
}
