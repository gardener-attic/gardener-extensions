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
	"github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack"
	openstacktypes "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/openstack"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	"github.com/gardener/gardener-extensions/pkg/util"
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

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	namespace = "test"
	authURL   = "someurl"
)

var dhcpDomain = util.StringPtr("dhcp-domain")
var requestTimeout = util.StringPtr("2s")

func defaultControlPlane() *extensionsv1alpha1.ControlPlane {
	return controlPlane(
		"floating-network-id",
		&openstack.ControlPlaneConfig{
			LoadBalancerProvider: "load-balancer-provider",
			CloudControllerManager: &openstack.CloudControllerManagerConfig{
				FeatureGates: map[string]bool{
					"CustomResourceValidation": true,
				},
			},
		})
}
func controlPlane(fnid string, cfg *openstack.ControlPlaneConfig) *extensionsv1alpha1.ControlPlane {
	return &extensionsv1alpha1.ControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "control-plane",
			Namespace: namespace,
		},
		Spec: extensionsv1alpha1.ControlPlaneSpec{
			SecretRef: corev1.SecretReference{
				Name:      common.CloudProviderSecretName,
				Namespace: namespace,
			},
			ProviderConfig: &runtime.RawExtension{
				Raw: encode(cfg),
			},
			InfrastructureProviderStatus: &runtime.RawExtension{
				Raw: encode(&openstack.InfrastructureStatus{
					Networks: openstack.NetworkStatus{
						FloatingPool: openstack.FloatingPoolStatus{
							ID: fnid,
						},
						Subnets: []openstack.Subnet{
							{
								ID:      "subnet-acbd1234",
								Purpose: openstack.PurposeNodes,
							},
						},
					},
				}),
			},
		},
	}
}

var _ = Describe("ValuesProvider", func() {
	var (
		ctrl *gomock.Controller

		// Build scheme
		scheme = runtime.NewScheme()
		_      = openstack.AddToScheme(scheme)

		cp = defaultControlPlane()

		cidr    = gardencorev1alpha1.CIDR("10.250.0.0/19")
		cluster = &extensionscontroller.Cluster{
			CloudProfile: &gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					OpenStack: &gardenv1beta1.OpenStackProfile{
						KeyStoneURL:    authURL,
						DHCPDomain:     dhcpDomain,
						RequestTimeout: requestTimeout,
					},
				},
			},
			Shoot: &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						OpenStack: &gardenv1beta1.OpenStackCloud{
							Networks: gardenv1beta1.OpenStackNetworks{
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

		cpSecretKey = client.ObjectKey{Namespace: namespace, Name: common.CloudProviderSecretName}
		cpSecret    = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      common.CloudProviderSecretName,
				Namespace: namespace,
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"domainName": []byte(`domain-name`),
				"tenantName": []byte(`tenant-name`),
				"username":   []byte(`username`),
				"password":   []byte(`password`),
			},
		}

		checksums = map[string]string{
			common.CloudProviderSecretName:         "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			openstacktypes.CloudProviderConfigName: "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606",
			"cloud-controller-manager":             "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
			"cloud-controller-manager-server":      "6dff2a2e6f14444b66d8e4a351c049f7e89ee24ba3eaab95dbec40ba6bdebb52",
		}

		configChartValues = map[string]interface{}{
			"kubernetesVersion": "1.13.4",
			"domainName":        "domain-name",
			"tenantName":        "tenant-name",
			"username":          "username",
			"password":          "password",
			"subnetID":          "subnet-acbd1234",
			"lbProvider":        "load-balancer-provider",
			"floatingNetworkID": "floating-network-id",
			"authUrl":           authURL,
			"dhcpDomain":        dhcpDomain,
			"requestTimeout":    requestTimeout,
		}

		ccmChartValues = map[string]interface{}{
			"replicas":          1,
			"kubernetesVersion": "1.13.4",
			"clusterName":       namespace,
			"podNetwork":        cidr,
			"podAnnotations": map[string]interface{}{
				"checksum/secret-cloud-controller-manager":        "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
				"checksum/secret-cloud-controller-manager-server": "6dff2a2e6f14444b66d8e4a351c049f7e89ee24ba3eaab95dbec40ba6bdebb52",
				"checksum/secret-cloudprovider":                   "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
				"checksum/configmap-cloud-provider-config":        "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606",
			},
			"featureGates": map[string]bool{
				"CustomResourceValidation": true,
			},
		}

		logger = log.Log.WithName("test")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#GetConfigChartValues", func() {
		It("should return correct config chart values", func() {
			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), cpSecretKey, &corev1.Secret{}).DoAndReturn(clientGet(cpSecret))

			// Create valuesProvider
			vp := NewValuesProvider(logger)
			err := vp.(inject.Scheme).InjectScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			err = vp.(inject.Client).InjectClient(client)
			Expect(err).NotTo(HaveOccurred())

			// Call GetConfigChartValues method and check the result
			values, err := vp.GetConfigChartValues(context.TODO(), cp, cluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(configChartValues))
		})
	})

	Describe("#GetControlPlaneChartValues", func() {
		It("should return correct control plane chart values", func() {
			// Create valuesProvider
			vp := NewValuesProvider(logger)
			err := vp.(inject.Scheme).InjectScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			// Call GetControlPlaneChartValues method and check the result
			values, err := vp.GetControlPlaneChartValues(context.TODO(), cp, cluster, checksums, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(ccmChartValues))
		})
	})

	Describe("#GetConfigChartValues with Classes", func() {
		It("should return correct config chart values", func() {
			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), cpSecretKey, &corev1.Secret{}).DoAndReturn(clientGet(cpSecret))

			// Create valuesProvider
			vp := NewValuesProvider(logger)
			err := vp.(inject.Scheme).InjectScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			err = vp.(inject.Client).InjectClient(client)
			Expect(err).NotTo(HaveOccurred())

			fnid := "4711"
			fnid2 := "pub"
			fsid := "0815"
			fsid2 := "pub0815"
			psid := "priv"
			dfsid := "default-floating-subnet-id"
			cp := controlPlane(
				fnid,
				&openstack.ControlPlaneConfig{
					LoadBalancerProvider: "load-balancer-provider",
					LoadBalancerClasses: []openstack.LoadBalancerClass{
						{
							Name:             "test",
							FloatingSubnetID: &fsid,
							SubnetID:         nil,
						},
						{
							Name:             "default",
							FloatingSubnetID: &dfsid,
							SubnetID:         nil,
						},
						{
							Name:              "public",
							FloatingSubnetID:  &fsid2,
							FloatingNetworkID: &fnid2,
							SubnetID:          nil,
						},
						{
							Name:     "other",
							SubnetID: &psid,
						},
					},
					CloudControllerManager: &openstack.CloudControllerManagerConfig{
						FeatureGates: map[string]bool{
							"CustomResourceValidation": true,
						},
					},
				},
			)

			configChartValues = map[string]interface{}{
				"kubernetesVersion": "1.13.4",
				"domainName":        "domain-name",
				"tenantName":        "tenant-name",
				"username":          "username",
				"password":          "password",
				"subnetID":          "subnet-acbd1234",
				"lbProvider":        "load-balancer-provider",
				"floatingNetworkID": fnid,
				"floatingSubnetID":  dfsid,
				"floatingClasses": []map[string]interface{}{
					{
						"name":              "test",
						"floatingNetworkID": fnid,
						"floatingSubnetID":  fsid,
					},
					{
						"name":              "default",
						"floatingNetworkID": fnid,
						"floatingSubnetID":  dfsid,
					},
					{
						"name":              "public",
						"floatingNetworkID": fnid2,
						"floatingSubnetID":  fsid2,
					},
					{
						"name":     "other",
						"subnetID": psid,
					},
				},
				"authUrl":        authURL,
				"dhcpDomain":     dhcpDomain,
				"requestTimeout": requestTimeout,
			}
			// Call GetConfigChartValues method and check the result
			values, err := vp.GetConfigChartValues(context.TODO(), cp, cluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(configChartValues))
		})
	})
})

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}

func clientGet(result runtime.Object) interface{} {
	return func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
		switch obj.(type) {
		case *corev1.Secret:
			*obj.(*corev1.Secret) = *result.(*corev1.Secret)
		}
		return nil
	}
}
