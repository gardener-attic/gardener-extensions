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

package internal

import (
	"encoding/json"
	"testing"

	apisaws "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	mockruntime "github.com/gardener/gardener-extensions/pkg/mock/apimachinery/runtime"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	namespace = "test"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS Controlplane Internal Suite")
}

var _ = Describe("Helper", func() {
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
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#ConfigChartValuesFunc", func() {
		It("should return correct configuration chart values", func() {
			var (
				expectedValues = map[string]interface{}{
					"cloudProviderConfig": `[Global]
VPC="vpc-1234"
SubnetID="subnet-acbd1234"
DisableSecurityGroupIngress=true
KubernetesClusterTag="` + namespace + `"
KubernetesClusterID="` + namespace + `"
Zone="eu-west-1a"
`,
				}
			)

			// Create mock decoder
			decoder := mockruntime.NewMockDecoder(ctrl)
			decoder.EXPECT().Decode(cp.Spec.ProviderConfig.Raw, nil, &apisaws.ControlPlaneConfig{}).DoAndReturn(decode)
			decoder.EXPECT().Decode(cp.Spec.InfrastructureProviderStatus.Raw, nil, &apisaws.InfrastructureStatus{}).DoAndReturn(decode)

			// Create helper
			h := NewHelper()

			// Call ConfigChartValuesFunc and check the results
			values, err := h.ConfigChartValuesFunc(decoder, cp, cluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(expectedValues))
		})
	})

	Describe("#ControlPlaneChartValuesFunc", func() {
		It("should return correct control plane chart values", func() {
			var (
				checksums = map[string]string{
					cloudControllerManagerDeploymentName: "1",
					cloudControllerManagerServerName:     "2",
					common.CloudProviderSecretName:       "3",
					common.CloudProviderConfigName:       "4",
				}

				expectedValues = map[string]interface{}{
					"cloudProvider":     "aws",
					"clusterName":       namespace,
					"kubernetesVersion": "1.13.4",
					"podNetwork":        cidr,
					"replicas":          1,
					"podAnnotations": map[string]interface{}{
						"checksum/secret-cloud-controller-manager":        "1",
						"checksum/secret-cloud-controller-manager-server": "2",
						"checksum/secret-cloudprovider":                   "3",
						"checksum/configmap-cloud-provider-config":        "4",
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

			// Create mock decoder
			decoder := mockruntime.NewMockDecoder(ctrl)
			decoder.EXPECT().Decode(cp.Spec.ProviderConfig.Raw, nil, &apisaws.ControlPlaneConfig{}).DoAndReturn(decode)

			// Create helper
			h := NewHelper()

			// Call ControlPlaneChartValuesFunc and check the results
			values, err := h.ControlPlaneChartValuesFunc(decoder, cp, cluster, checksums)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(expectedValues))
		})
	})

})

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}

func decode(data []byte, _ *schema.GroupVersionKind, obj runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	var err error
	switch x := obj.(type) {
	case *apisaws.ControlPlaneConfig:
		err = json.Unmarshal(data, x)
	case *apisaws.InfrastructureStatus:
		err = json.Unmarshal(data, x)
	}
	return nil, nil, err
}
