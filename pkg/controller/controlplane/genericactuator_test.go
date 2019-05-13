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

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockcontrolplane "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/controller/controlplane"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/gardener/gardener/pkg/utils/imagevector"

	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	namespace               = "test"
	cloudProviderConfigName = "cloud-provider-config"
)

var _ = Describe("Actuator", func() {
	var (
		ctrl *gomock.Controller

		cp = &extensionsv1alpha1.ControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "control-plane",
				Namespace: namespace,
			},
			Spec: extensionsv1alpha1.ControlPlaneSpec{},
		}
		cluster = &extensionscontroller.Cluster{
			Shoot: &gardenv1beta1.Shoot{},
		}

		deployedSecrets = map[string]*corev1.Secret{
			"cloud-controller-manager": {
				ObjectMeta: metav1.ObjectMeta{Name: "cloud-controller-manager", Namespace: namespace},
				Data:       map[string][]byte{"a": []byte("b")},
			},
		}

		cpSecretKey    = client.ObjectKey{Namespace: namespace, Name: common.CloudProviderSecretName}
		cpConfigMapKey = client.ObjectKey{Namespace: namespace, Name: cloudProviderConfigName}
		cpSecret       = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      common.CloudProviderSecretName,
				Namespace: namespace,
			},
			Data: map[string][]byte{"foo": []byte("bar")},
		}
		cpConfigMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cloudProviderConfigName,
				Namespace: namespace,
			},
			Data: map[string]string{"abc": "xyz"},
		}

		imageVector = imagevector.ImageVector([]*imagevector.ImageSource{})

		checksums = map[string]string{
			common.CloudProviderSecretName: "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			cloudProviderConfigName:        "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606",
			"cloud-controller-manager":     "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
		}
		checksumsWithoutConfig = map[string]string{
			common.CloudProviderSecretName: "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			"cloud-controller-manager":     "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
		}

		configChartValues = map[string]interface{}{
			"cloudProviderConfig": `[Global]`,
		}

		controlPlaneChartValues = map[string]interface{}{
			"clusterName": namespace,
		}

		logger = log.Log.WithName("test")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#Reconcile", func() {
		It("should deploy secrets and apply charts with correct parameters", func() {
			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), cpSecretKey, &corev1.Secret{}).DoAndReturn(clientGet(cpSecret))
			client.EXPECT().Get(context.TODO(), cpConfigMapKey, &corev1.ConfigMap{}).DoAndReturn(clientGet(cpConfigMap))

			// Create mock secrets and charts
			secrets := mockcontrolplane.NewMockSecrets(ctrl)
			secrets.EXPECT().Deploy(gomock.Any(), gomock.Any(), namespace).Return(deployedSecrets, nil)
			configChart := mockcontrolplane.NewMockChart(ctrl)
			configChart.EXPECT().Apply(context.TODO(), gomock.Any(), gomock.Any(), namespace, cluster.Shoot, nil, nil, configChartValues).Return(nil)
			ccmChart := mockcontrolplane.NewMockChart(ctrl)
			ccmChart.EXPECT().Apply(context.TODO(), gomock.Any(), gomock.Any(), namespace, cluster.Shoot, imageVector, checksums, controlPlaneChartValues).Return(nil)

			// Create mock values provider
			vp := mockcontrolplane.NewMockValuesProvider(ctrl)
			vp.EXPECT().GetConfigChartValues(context.TODO(), cp, cluster).Return(configChartValues, nil)
			vp.EXPECT().GetControlPlaneChartValues(context.TODO(), cp, cluster, checksums).Return(controlPlaneChartValues, nil)

			// Create actuator
			a := NewActuator(secrets, configChart, ccmChart, vp, imageVector, cloudProviderConfigName, logger)
			err := a.(inject.Client).InjectClient(client)
			Expect(err).NotTo(HaveOccurred())

			// Call Reconcile method and check the result
			err = a.Reconcile(context.TODO(), cp, cluster)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should deploy secrets and apply charts with correct parameters (only controlplane chart)", func() {
			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), cpSecretKey, &corev1.Secret{}).DoAndReturn(clientGet(cpSecret))

			// Create mock secrets and charts
			secrets := mockcontrolplane.NewMockSecrets(ctrl)
			secrets.EXPECT().Deploy(gomock.Any(), gomock.Any(), namespace).Return(deployedSecrets, nil)
			ccmChart := mockcontrolplane.NewMockChart(ctrl)
			ccmChart.EXPECT().Apply(context.TODO(), gomock.Any(), gomock.Any(), namespace, cluster.Shoot, imageVector, checksumsWithoutConfig, controlPlaneChartValues).Return(nil)

			// Create mock values provider
			vp := mockcontrolplane.NewMockValuesProvider(ctrl)
			vp.EXPECT().GetControlPlaneChartValues(context.TODO(), cp, cluster, checksumsWithoutConfig).Return(controlPlaneChartValues, nil)

			// Create actuator
			a := NewActuator(secrets, nil, ccmChart, vp, imageVector, "", logger)
			err := a.(inject.Client).InjectClient(client)
			Expect(err).NotTo(HaveOccurred())

			// Call Reconcile method and check the result
			err = a.Reconcile(context.TODO(), cp, cluster)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("#Delete", func() {
		It("should delete secrets and charts", func() {
			// Create mock client
			client := mockclient.NewMockClient(ctrl)

			// Create mock secrets and charts
			secrets := mockcontrolplane.NewMockSecrets(ctrl)
			secrets.EXPECT().Delete(gomock.Any(), namespace).Return(nil)
			configChart := mockcontrolplane.NewMockChart(ctrl)
			configChart.EXPECT().Delete(context.TODO(), client, namespace).Return(nil)
			ccmChart := mockcontrolplane.NewMockChart(ctrl)
			ccmChart.EXPECT().Delete(context.TODO(), client, namespace).Return(nil)

			// Create actuator
			a := NewActuator(secrets, configChart, ccmChart, nil, nil, cloudProviderConfigName, logger)
			err := a.(inject.Client).InjectClient(client)
			Expect(err).NotTo(HaveOccurred())

			// Call Delete method and check the result
			err = a.Delete(context.TODO(), cp, cluster)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete secrets and charts (only controlplane chart)", func() {
			// Create mock client
			client := mockclient.NewMockClient(ctrl)

			// Create mock secrets and charts
			secrets := mockcontrolplane.NewMockSecrets(ctrl)
			secrets.EXPECT().Delete(gomock.Any(), namespace).Return(nil)
			ccmChart := mockcontrolplane.NewMockChart(ctrl)
			ccmChart.EXPECT().Delete(context.TODO(), client, namespace).Return(nil)

			// Create actuator
			a := NewActuator(secrets, nil, ccmChart, nil, nil, "", logger)
			err := a.(inject.Client).InjectClient(client)
			Expect(err).NotTo(HaveOccurred())

			// Call Delete method and check the result
			err = a.Delete(context.TODO(), cp, cluster)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func clientGet(result runtime.Object) interface{} {
	return func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
		switch obj.(type) {
		case *corev1.Secret:
			*obj.(*corev1.Secret) = *result.(*corev1.Secret)
		case *corev1.ConfigMap:
			*obj.(*corev1.ConfigMap) = *result.(*corev1.ConfigMap)
		}
		return nil
	}
}
