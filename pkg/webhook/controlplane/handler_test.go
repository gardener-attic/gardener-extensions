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
	"errors"
	"net/http"
	"testing"

	mockmanager "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/manager"
	mocktypes "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/webhook/admission/types"
	mockcontrolplane "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/webhook/controlplane"

	"github.com/appscode/jsonpatch"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

func TestControlplane(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controlplane Webhook Suite")
}

var _ = Describe("Handler", func() {
	const (
		name      = "foo"
		namespace = "default"
	)

	var (
		ctrl    *gomock.Controller
		mgr     *mockmanager.MockManager
		decoder *mocktypes.MockDecoder

		objTypes = []runtime.Object{&corev1.Service{}}
		svc      = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		}

		req = types.Request{
			AdmissionRequest: &admissionv1beta1.AdmissionRequest{
				Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
				Name:      name,
				Namespace: namespace,
				Operation: admissionv1beta1.Create,
			},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		// Build scheme
		scheme := runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)

		// Create mock manager
		mgr = mockmanager.NewMockManager(ctrl)
		mgr.EXPECT().GetScheme().Return(scheme)

		// Create mock decoder
		decoder = mocktypes.NewMockDecoder(ctrl)
		decoder.EXPECT().Decode(req, &corev1.Service{}).DoAndReturn(decoderDecode(svc))
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#Handle", func() {
		It("should return an allowing response if the resource wasn't changed by mutator", func() {
			// Create mock mutator
			mutator := mockcontrolplane.NewMockMutator(ctrl)
			mutator.EXPECT().Mutate(context.TODO(), svc).Return(nil)

			// Create handler
			h, err := newHandler(mgr, objTypes, mutator, logger)
			Expect(err).NotTo(HaveOccurred())
			h.decoder = decoder

			// Call Handle and check response
			resp := h.Handle(context.TODO(), req)
			Expect(resp).To(Equal(types.Response{
				Response: &admissionv1beta1.AdmissionResponse{
					Allowed: true,
				},
			}))
		})

		It("should return a patch response if the resource was changed by mutator", func() {
			// Create mock mutator
			mutator := mockcontrolplane.NewMockMutator(ctrl)
			mutator.EXPECT().Mutate(context.TODO(), svc).DoAndReturn(func(ctx context.Context, obj runtime.Object) error {
				accessor, _ := meta.Accessor(obj)
				accessor.SetAnnotations(map[string]string{"foo": "bar"})
				return nil
			})

			// Create handler
			h, err := newHandler(mgr, objTypes, mutator, logger)
			Expect(err).NotTo(HaveOccurred())
			h.decoder = decoder

			// Call Handle and check response
			resp := h.Handle(context.TODO(), req)
			pt := admissionv1beta1.PatchTypeJSONPatch
			Expect(resp).To(Equal(types.Response{
				Patches: []jsonpatch.JsonPatchOperation{
					{
						Operation: "add",
						Path:      "/metadata/annotations",
						Value:     map[string]interface{}{"foo": "bar"},
					},
				},
				Response: &admissionv1beta1.AdmissionResponse{
					Allowed:   true,
					PatchType: &pt,
				},
			}))
		})

		It("should return an error response if the mutator returned an error", func() {
			// Create mock mutator
			mutator := mockcontrolplane.NewMockMutator(ctrl)
			mutator.EXPECT().Mutate(context.TODO(), svc).Return(errors.New("test error"))

			// Create handler
			h, err := newHandler(mgr, objTypes, mutator, logger)
			Expect(err).NotTo(HaveOccurred())
			h.decoder = decoder

			// Call Handle and check response
			resp := h.Handle(context.TODO(), req)
			Expect(resp).To(Equal(types.Response{
				Response: &admissionv1beta1.AdmissionResponse{
					Allowed: false,
					Result: &metav1.Status{
						Code:    http.StatusInternalServerError,
						Message: "could not mutate Service default/foo: test error",
					},
				},
			}))
		})
	})
})

func decoderDecode(result runtime.Object) interface{} {
	return func(ar types.Request, obj runtime.Object) error {
		switch obj.(type) {
		case *corev1.Service:
			*obj.(*corev1.Service) = *result.(*corev1.Service)
		}
		return nil
	}
}
