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
	"encoding/json"
	"testing"

	apisaws "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/imagevector"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	mockruntime "github.com/gardener/gardener-extensions/pkg/mock/apimachinery/runtime"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockcontrolplane "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/controller/controlplane"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

		cp = &extensionsv1alpha1.ControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "control-plane",
				Namespace: namespace,
			},
			Spec: extensionsv1alpha1.ControlPlaneSpec{
				ProviderConfig: &runtime.RawExtension{
					Raw: encode(&apisaws.ControlPlaneConfig{
						CloudControllerManager: &apisaws.CloudControllerManagerConfig{
							KubernetesConfig: gardenv1beta1.KubernetesConfig{
								FeatureGates: map[string]bool{
									"CustomResourceValidation": true,
								},
							},
						},
					}),
				},
				InfrastructureProviderStatus: &runtime.RawExtension{
					Raw: encode(&apisaws.InfrastructureStatus{
						VPC: apisaws.VPCStatus{
							ID: "vpc-1234",
							Subnets: []apisaws.Subnet{
								{
									ID:      "subnet-acbd1234",
									Purpose: "public",
									Zone:    "eu-west-1a",
								},
							},
						},
					}),
				},
			},
		}

		cidr    = gardencorev1alpha1.CIDR("10.250.0.0/19")
		cluster = &extensionscontroller.Cluster{
			Shoot: &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						AWS: &gardenv1beta1.AWSCloud{
							Networks: gardenv1beta1.AWSNetworks{
								K8SNetworks: gardencorev1alpha1.K8SNetworks{
									Pods: &cidr,
								},
							},
						},
					},
					Kubernetes: gardenv1beta1.Kubernetes{
						Version: "1.13.4",
					},
				},
			},
		}

		deployedSecrets = map[string]*corev1.Secret{
			"cloud-controller-manager": {
				ObjectMeta: metav1.ObjectMeta{Name: "cloud-controller-manager", Namespace: namespace},
				Data:       map[string][]byte{"a": []byte("b")},
			},
			"cloud-controller-manager-server": {
				ObjectMeta: metav1.ObjectMeta{Name: "cloud-controller-manager-server", Namespace: namespace},
				Data:       map[string][]byte{"c": []byte("d")},
			},
		}

		cpSecretKey    = client.ObjectKey{Namespace: namespace, Name: common.CloudProviderSecretName}
		cpConfigMapKey = client.ObjectKey{Namespace: namespace, Name: aws.CloudProviderConfigName}
		cpSecret       = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      common.CloudProviderSecretName,
				Namespace: namespace,
			},
			Data: map[string][]byte{"foo": []byte("bar")},
		}
		cpConfigMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      aws.CloudProviderConfigName,
				Namespace: namespace,
			},
			Data: map[string]string{"abc": "xyz"},
		}

		imageVector = imagevector.ImageVector()

		checksums = map[string]string{
			common.CloudProviderSecretName:    "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			aws.CloudProviderConfigName:       "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606",
			"cloud-controller-manager":        "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
			"cloud-controller-manager-server": "6dff2a2e6f14444b66d8e4a351c049f7e89ee24ba3eaab95dbec40ba6bdebb52",
		}

		configChartValues = map[string]interface{}{
			"cloudProviderConfig": `[Global]
VPC="vpc-1234"
SubnetID="subnet-acbd1234"
DisableSecurityGroupIngress=true
KubernetesClusterTag="` + namespace + `"
KubernetesClusterID="` + namespace + `"
Zone="eu-west-1a"
`,
		}

		ccmValues = map[string]interface{}{
			"cloudProvider":     "aws",
			"clusterName":       namespace,
			"kubernetesVersion": "1.13.4",
			"podNetwork":        cidr,
			"replicas":          1,
			"podAnnotations": map[string]interface{}{
				"checksum/secret-cloud-controller-manager":        "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
				"checksum/secret-cloud-controller-manager-server": "6dff2a2e6f14444b66d8e4a351c049f7e89ee24ba3eaab95dbec40ba6bdebb52",
				"checksum/secret-cloudprovider":                   "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
				"checksum/configmap-cloud-provider-config":        "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606",
			},
			"configureRoutes": false,
			"environment": []map[string]interface{}{
				{
					"name": "AWS_ACCESS_KEY_ID",
					"valueFrom": map[string]interface{}{
						"secretKeyRef": map[string]interface{}{
							"key":  aws.AccessKeyID,
							"name": common.CloudProviderSecretName,
						},
					},
				},
				{
					"name": "AWS_SECRET_ACCESS_KEY",
					"valueFrom": map[string]interface{}{
						"secretKeyRef": map[string]interface{}{
							"key":  aws.SecretAccessKey,
							"name": common.CloudProviderSecretName,
						},
					},
				},
			},
			"resources": map[string]interface{}{
				"limits": map[string]interface{}{
					"cpu":    "500m",
					"memory": "512Mi",
				},
			},
			"featureGates": map[string]bool{
				"CustomResourceValidation": true,
			},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#Reconcile", func() {
		It("should deploy secrets and apply charts with correct parameters", func() {
			// Create mock decoder
			decoder := mockruntime.NewMockDecoder(ctrl)
			decoder.EXPECT().Decode(cp.Spec.ProviderConfig.Raw, nil, &apisaws.ControlPlaneConfig{}).DoAndReturn(decoderDecode())
			decoder.EXPECT().Decode(cp.Spec.InfrastructureProviderStatus.Raw, nil, &apisaws.InfrastructureStatus{}).DoAndReturn(decoderDecode())

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
			ccmChart.EXPECT().Apply(context.TODO(), gomock.Any(), gomock.Any(), namespace, cluster.Shoot, imageVector, checksums, ccmValues).Return(nil)

			// Create actuator
			a := NewActuator(secrets, configChart, ccmChart)
			a.(*actuator).decoder = decoder
			a.(*actuator).client = client

			// Call Reconcile method and check the result
			err := a.Reconcile(context.TODO(), cp, cluster)
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
			a := NewActuator(secrets, configChart, ccmChart)
			a.(*actuator).client = client

			// Call Delete method and check the result
			err := a.Delete(context.TODO(), cp, cluster)
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

func decoderDecode() interface{} {
	return func(data []byte, _ *schema.GroupVersionKind, obj runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
		err := json.Unmarshal(data, obj)
		return nil, nil, err
	}
}

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}
