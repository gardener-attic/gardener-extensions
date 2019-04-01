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

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/controlplane/internal"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockcontrolplane "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/controller/controlplane"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace = "test"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS Controlplane Suite")
}

var _ = Describe("Actuator", func() {
	var (
		ctrl *gomock.Controller
		cp   = &extensionsv1alpha1.ControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "control-plane",
				Namespace: namespace,
			},
		}
		cluster = &extensionscontroller.Cluster{}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#Reconcile", func() {
		It("should deploy secrets and apply charts with correct parameters", func() {
			var (
				imageVector = imagevector.ImageVector([]*imagevector.ImageSource{})

				configChartValues = map[string]interface{}{
					"foo": "bar",
				}
				controlPlaneChartValues = map[string]interface{}{
					"abc": "xyz",
				}

				cpSecretKey    = client.ObjectKey{Namespace: namespace, Name: common.CloudProviderSecretName}
				cpConfigMapKey = client.ObjectKey{Namespace: namespace, Name: common.CloudProviderConfigName}

				cpSecret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      common.CloudProviderSecretName,
						Namespace: namespace,
					},
					Data: map[string][]byte{"foo": []byte("bar")},
				}
				cpConfigMap = &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      common.CloudProviderConfigName,
						Namespace: namespace,
					},
					Data: map[string]string{"abc": "xyz"},
				}

				checksums = map[string]string{
					common.CloudProviderSecretName: "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
					common.CloudProviderConfigName: "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606",
				}
			)

			// Create mock client
			c := mockclient.NewMockClient(ctrl)
			c.EXPECT().Get(context.TODO(), cpSecretKey, &corev1.Secret{}).DoAndReturn(
				func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
					*obj.(*corev1.Secret) = *cpSecret
					return nil
				},
			)
			c.EXPECT().Get(context.TODO(), cpConfigMapKey, &corev1.ConfigMap{}).DoAndReturn(
				func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
					*obj.(*corev1.ConfigMap) = *cpConfigMap
					return nil
				},
			)

			// Create mock secrets
			secrets := mockcontrolplane.NewMockSecrets(ctrl)
			secrets.EXPECT().Deploy(gomock.Any(), gomock.Any(), namespace).Return(nil, nil)

			// Create mock configuration chart
			configChart := mockcontrolplane.NewMockChart(ctrl)
			configChart.EXPECT().Apply(context.TODO(), gomock.Any(), gomock.Any(), namespace, nil, nil, nil, configChartValues).Return(nil)

			// Create mock control plane chart
			controlPlaneChart := mockcontrolplane.NewMockChart(ctrl)
			controlPlaneChart.EXPECT().Apply(context.TODO(), gomock.Any(), gomock.Any(), namespace, nil, imageVector, checksums, controlPlaneChartValues).Return(nil)

			// Create actuator helper
			helper := &internal.Helper{
				Secrets:           secrets,
				ConfigChart:       configChart,
				ControlPlaneChart: controlPlaneChart,
				ImageVectorFunc: func() imagevector.ImageVector {
					return imageVector
				},
				ConfigChartValuesFunc: func(runtime.Decoder, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster) (map[string]interface{}, error) {
					return configChartValues, nil
				},
				ControlPlaneChartValuesFunc: func(runtime.Decoder, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster, map[string]string) (map[string]interface{}, error) {
					return controlPlaneChartValues, nil
				},
			}

			// Create actuator
			a := NewActuator(helper)
			err := a.(*actuator).InjectClient(c)
			Expect(err).NotTo(HaveOccurred())

			// Call Reconcile method and check the result
			err = a.Reconcile(context.TODO(), cp, cluster)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("#Delete", func() {
		It("should delete secrets and charts", func() {
			// Create mock secrets
			secrets := mockcontrolplane.NewMockSecrets(ctrl)
			secrets.EXPECT().Delete(gomock.Any(), namespace).Return(nil)

			// Create mock configuration chart
			configChart := mockcontrolplane.NewMockChart(ctrl)
			configChart.EXPECT().Delete(context.TODO(), gomock.Any(), namespace).Return(nil)

			// Create mock control plane chart
			controlPlaneChart := mockcontrolplane.NewMockChart(ctrl)
			controlPlaneChart.EXPECT().Delete(context.TODO(), gomock.Any(), namespace).Return(nil)

			// Create actuator helper
			helper := &internal.Helper{
				Secrets:           secrets,
				ConfigChart:       configChart,
				ControlPlaneChart: controlPlaneChart,
			}

			// Create actuator
			a := NewActuator(helper)

			// Call Delete method and check the result
			err := a.Delete(context.TODO(), cp, cluster)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
