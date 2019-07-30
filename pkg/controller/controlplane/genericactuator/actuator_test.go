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

package genericactuator

import (
	"context"
	"testing"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockextensionscontroller "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/controller"
	mockgenericactuator "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/controller/controlplane/genericactuator"
	mockutil "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/util"
	mockchartrenderer "github.com/gardener/gardener-extensions/pkg/mock/gardener/chartrenderer"
	mockkubernetes "github.com/gardener/gardener-extensions/pkg/mock/gardener/client/kubernetes"
	"github.com/gardener/gardener-extensions/pkg/util"

	resourcemanagerv1alpha1 "github.com/gardener/gardener-resource-manager/pkg/apis/resources/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener-resource-manager/pkg/apis/resources/v1alpha1"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	namespace               = "test"
	cloudProviderConfigName = "cloud-provider-config"
	chartName               = "chartName"
	renderedContent         = "renderedContent"

	seedVersion  = "1.13.0"
	shootVersion = "1.14.0"
)

func TestControlplane(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controlplane Generic Actuator Suite")
}

var _ = Describe("Actuator", func() {
	var (
		ctrl *gomock.Controller

		cp = &extensionsv1alpha1.ControlPlane{
			ObjectMeta: metav1.ObjectMeta{Name: "control-plane", Namespace: namespace},
			Spec:       extensionsv1alpha1.ControlPlaneSpec{},
		}
		cluster = &extensionscontroller.Cluster{
			Shoot: &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Kubernetes: gardenv1beta1.Kubernetes{
						Version: shootVersion,
					},
				},
			},
		}

		deployedSecrets = map[string]*corev1.Secret{
			"cloud-controller-manager": {
				ObjectMeta: metav1.ObjectMeta{Name: "cloud-controller-manager", Namespace: namespace},
				Data:       map[string][]byte{"a": []byte("b")},
			},
		}

		cpSecretKey    = client.ObjectKey{Namespace: namespace, Name: gardencorev1alpha1.SecretNameCloudProvider}
		cpConfigMapKey = client.ObjectKey{Namespace: namespace, Name: cloudProviderConfigName}
		cpSecret       = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: gardencorev1alpha1.SecretNameCloudProvider, Namespace: namespace},
			Data:       map[string][]byte{"foo": []byte("bar")},
		}
		cpConfigMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: cloudProviderConfigName, Namespace: namespace},
			Data:       map[string]string{"abc": "xyz"},
		}

		resourceKeyCPShootChart        = client.ObjectKey{Namespace: namespace, Name: controlPlaneShootChartResourceName}
		createdMRSecretForCPShootChart = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: controlPlaneShootChartResourceName, Namespace: namespace},
			Data:       map[string][]byte{chartName: []byte(renderedContent)},
			Type:       corev1.SecretTypeOpaque,
		}
		createdMRForCPShootChart = &resourcemanagerv1alpha1.ManagedResource{
			ObjectMeta: metav1.ObjectMeta{Name: controlPlaneShootChartResourceName, Namespace: namespace},
			Spec: resourcemanagerv1alpha1.ManagedResourceSpec{
				SecretRefs: []corev1.LocalObjectReference{
					{Name: controlPlaneShootChartResourceName},
				},
				InjectLabels: map[string]string{extensionscontroller.ShootNoCleanupLabel: "true"},
			},
		}
		deletedMRSecretForCPShootChart = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: controlPlaneShootChartResourceName, Namespace: namespace},
			Type:       corev1.SecretTypeOpaque,
		}
		deleteMRForCPShootChart = &resourcemanagerv1alpha1.ManagedResource{
			ObjectMeta: metav1.ObjectMeta{Name: controlPlaneShootChartResourceName, Namespace: namespace},
			Spec: resourcemanagerv1alpha1.ManagedResourceSpec{
				SecretRefs:   []corev1.LocalObjectReference{},
				InjectLabels: map[string]string{},
			},
		}

		resourceKeyStorageClassesChart        = client.ObjectKey{Namespace: namespace, Name: storageClassesChartResourceName}
		createdMRSecretForStorageClassesChart = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: storageClassesChartResourceName, Namespace: namespace},
			Data:       map[string][]byte{chartName: []byte(renderedContent)},
			Type:       corev1.SecretTypeOpaque,
		}
		createdMRForStorageClassesChart = &resourcemanagerv1alpha1.ManagedResource{
			ObjectMeta: metav1.ObjectMeta{Name: storageClassesChartResourceName, Namespace: namespace},
			Spec: resourcemanagerv1alpha1.ManagedResourceSpec{
				SecretRefs: []corev1.LocalObjectReference{
					{Name: storageClassesChartResourceName},
				},
				InjectLabels: map[string]string{extensionscontroller.ShootNoCleanupLabel: "true"},
			},
		}
		deletedMRSecretForStorageClassesChart = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: storageClassesChartResourceName, Namespace: namespace},
			Type:       corev1.SecretTypeOpaque,
		}
		deleteMRForStorageClassesChart = &resourcemanagerv1alpha1.ManagedResource{
			ObjectMeta: metav1.ObjectMeta{Name: storageClassesChartResourceName, Namespace: namespace},
			Spec: resourcemanagerv1alpha1.ManagedResourceSpec{
				SecretRefs:   []corev1.LocalObjectReference{},
				InjectLabels: map[string]string{},
			},
		}

		imageVector = imagevector.ImageVector([]*imagevector.ImageSource{})

		checksums = map[string]string{
			gardencorev1alpha1.SecretNameCloudProvider: "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			cloudProviderConfigName:                    "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606",
			"cloud-controller-manager":                 "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
		}
		checksumsNoConfig = map[string]string{
			gardencorev1alpha1.SecretNameCloudProvider: "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			"cloud-controller-manager":                 "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
		}

		configChartValues = map[string]interface{}{
			"cloudProviderConfig": `[Global]`,
		}

		controlPlaneChartValues = map[string]interface{}{
			"clusterName": namespace,
		}

		controlPlaneShootChartValues = map[string]interface{}{
			"foo": "bar",
		}

		storageClassesChartValues = map[string]interface{}{
			"foo": "bar",
		}

		errNotFound = &errors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
		logger      = log.Log.WithName("test")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	DescribeTable("#Reconcile",
		func(configName string, checksums map[string]string) {
			ctx := context.TODO()

			// Create mock client
			client := mockclient.NewMockClient(ctrl)

			client.EXPECT().Get(ctx, cpSecretKey, &corev1.Secret{}).DoAndReturn(clientGet(cpSecret))
			if configName != "" {
				client.EXPECT().Get(ctx, cpConfigMapKey, &corev1.ConfigMap{}).DoAndReturn(clientGet(cpConfigMap))
			}

			client.EXPECT().Get(ctx, resourceKeyCPShootChart, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(errNotFound)
			client.EXPECT().Create(ctx, createdMRSecretForCPShootChart.DeepCopy()).Return(nil)
			client.EXPECT().Get(ctx, resourceKeyCPShootChart, gomock.AssignableToTypeOf(&resourcesv1alpha1.ManagedResource{})).Return(errNotFound)
			client.EXPECT().Create(ctx, createdMRForCPShootChart.DeepCopy()).Return(nil)

			client.EXPECT().Get(ctx, resourceKeyStorageClassesChart, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(errNotFound)
			client.EXPECT().Create(ctx, createdMRSecretForStorageClassesChart.DeepCopy()).Return(nil)
			client.EXPECT().Get(ctx, resourceKeyStorageClassesChart, gomock.AssignableToTypeOf(&resourcesv1alpha1.ManagedResource{})).Return(errNotFound)
			client.EXPECT().Create(ctx, createdMRForStorageClassesChart.DeepCopy()).Return(nil)

			// Create mock Gardener clientset and chart applier
			gardenerClientset := mockkubernetes.NewMockInterface(ctrl)
			gardenerClientset.EXPECT().Version().Return(seedVersion)
			chartApplier := mockkubernetes.NewMockChartApplier(ctrl)

			// Create mock chart renderer and factory
			chartRenderer := mockchartrenderer.NewMockInterface(ctrl)
			crf := mockextensionscontroller.NewMockChartRendererFactory(ctrl)
			crf.EXPECT().NewChartRendererForShoot(shootVersion).Return(chartRenderer, nil)

			// Create mock secrets and charts
			secrets := mockutil.NewMockSecrets(ctrl)
			secrets.EXPECT().Deploy(gomock.Any(), gardenerClientset, namespace).Return(deployedSecrets, nil)
			var configChart util.Chart
			if configName != "" {
				cc := mockutil.NewMockChart(ctrl)
				cc.EXPECT().Apply(ctx, chartApplier, namespace, nil, "", "", configChartValues).Return(nil)
				configChart = cc
			}
			ccmChart := mockutil.NewMockChart(ctrl)
			ccmChart.EXPECT().Apply(ctx, chartApplier, namespace, imageVector, seedVersion, shootVersion, controlPlaneChartValues).Return(nil)
			ccmShootChart := mockutil.NewMockChart(ctrl)
			ccmShootChart.EXPECT().Render(chartRenderer, metav1.NamespaceSystem, imageVector, shootVersion, shootVersion, controlPlaneShootChartValues).Return(chartName, []byte(renderedContent), nil)
			storageClassesChart := mockutil.NewMockChart(ctrl)
			storageClassesChart.EXPECT().Render(chartRenderer, metav1.NamespaceSystem, imageVector, shootVersion, shootVersion, storageClassesChartValues).Return(chartName, []byte(renderedContent), nil)

			// Create mock values provider
			vp := mockgenericactuator.NewMockValuesProvider(ctrl)
			if configName != "" {
				vp.EXPECT().GetConfigChartValues(ctx, cp, cluster).Return(configChartValues, nil)
			}
			vp.EXPECT().GetControlPlaneChartValues(ctx, cp, cluster, checksums, false).Return(controlPlaneChartValues, nil)
			vp.EXPECT().GetControlPlaneShootChartValues(ctx, cp, cluster).Return(controlPlaneShootChartValues, nil)
			vp.EXPECT().GetStorageClassesChartValues(ctx, cp, cluster).Return(storageClassesChartValues, nil)

			// Create actuator
			a := NewActuator(secrets, configChart, ccmChart, ccmShootChart, storageClassesChart, vp, crf, imageVector, configName, logger)
			err := a.(inject.Client).InjectClient(client)
			Expect(err).NotTo(HaveOccurred())
			a.(*actuator).gardenerClientset = gardenerClientset
			a.(*actuator).chartApplier = chartApplier

			// Call Reconcile method and check the result
			requeue, err := a.Reconcile(ctx, cp, cluster)
			Expect(requeue).To(Equal(false))
			Expect(err).NotTo(HaveOccurred())
		},
		Entry("should deploy secrets and apply charts with correct parameters", cloudProviderConfigName, checksums),
		Entry("should deploy secrets and apply charts with correct parameters (no config)", "", checksumsNoConfig),
	)

	DescribeTable("#Delete",
		func(configName string) {
			ctx := context.TODO()

			// Create mock clients
			client := mockclient.NewMockClient(ctrl)

			client.EXPECT().Delete(ctx, deletedMRSecretForStorageClassesChart).Return(nil)
			client.EXPECT().Delete(ctx, deleteMRForStorageClassesChart).Return(nil)

			client.EXPECT().Delete(ctx, deletedMRSecretForCPShootChart).Return(nil)
			client.EXPECT().Delete(ctx, deleteMRForCPShootChart).Return(nil)

			// Create mock secrets and charts
			secrets := mockutil.NewMockSecrets(ctrl)
			secrets.EXPECT().Delete(gomock.Any(), namespace).Return(nil)
			var configChart util.Chart
			if configName != "" {
				cc := mockutil.NewMockChart(ctrl)
				cc.EXPECT().Delete(ctx, client, namespace).Return(nil)
				configChart = cc
			}
			ccmChart := mockutil.NewMockChart(ctrl)
			ccmChart.EXPECT().Delete(ctx, client, namespace).Return(nil)

			// Create actuator
			a := NewActuator(secrets, configChart, ccmChart, nil, nil, nil, nil, nil, configName, logger)
			err := a.(inject.Client).InjectClient(client)
			Expect(err).NotTo(HaveOccurred())

			// Call Delete method and check the result
			err = a.Delete(ctx, cp, cluster)
			Expect(err).NotTo(HaveOccurred())
		},
		Entry("should delete secrets and charts", cloudProviderConfigName),
		Entry("should delete secrets and charts (no config)", ""),
	)
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
