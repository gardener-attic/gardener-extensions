/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package controlplane

import (
	"context"
	"encoding/json"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane/genericactuator"

	apisvsphere "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere"
	"github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/vsphere"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	v1alpha1constants "github.com/gardener/gardener/pkg/apis/core/v1alpha1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

const (
	namespace = "shoot--foo--bar"
)

var _ = Describe("ValuesProvider", func() {
	var (
		ctrl *gomock.Controller

		// Build scheme
		scheme = runtime.NewScheme()
		_      = apisvsphere.AddToScheme(scheme)

		cpConfig = &apisvsphere.ControlPlaneConfig{
			CloudControllerManager: &apisvsphere.CloudControllerManagerConfig{
				FeatureGates: map[string]bool{
					"CustomResourceValidation": true,
				},
			},
			LoadBalancerClassNames: []string{"private"},
		}

		cp = &extensionsv1alpha1.ControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "control-plane",
				Namespace: namespace,
			},
			Spec: extensionsv1alpha1.ControlPlaneSpec{
				SecretRef: corev1.SecretReference{
					Name:      v1alpha1constants.SecretNameCloudProvider,
					Namespace: namespace,
				},
				ProviderConfig: &runtime.RawExtension{
					Raw: encode(cpConfig),
				},
				InfrastructureProviderStatus: &runtime.RawExtension{
					Raw: encode(&apisvsphere.InfrastructureStatus{
						Network: "gardener-test-network",
					}),
				},
			},
		}

		cidr    = "10.250.0.0/19"
		cluster = &extensionscontroller.Cluster{
			CloudProfile: &gardencorev1alpha1.CloudProfile{
				Spec: gardencorev1alpha1.CloudProfileSpec{
					ProviderConfig: &gardencorev1alpha1.ProviderConfig{
						RawExtension: runtime.RawExtension{
							Raw: encode(&apisvsphere.CloudProfileConfig{
								NamePrefix:                    "nameprefix",
								DefaultClassStoragePolicyName: "mypolicy",
								Regions: []apisvsphere.RegionSpec{
									{
										Name:               "testregion",
										VsphereHost:        "vsphere.host.internal",
										VsphereInsecureSSL: true,
										NSXTHost:           "nsxt.host.internal",
										NSXTInsecureSSL:    true,
										TransportZone:      "tz",
										LogicalTier0Router: "lt0router",
										EdgeCluster:        "edgecluster",
										SNATIPPool:         "snatIpPool",
										Datacenter:         "scc01-DC",
										Datastore:          "A800_VMwareB",
										Zones: []apisvsphere.ZoneSpec{
											{
												Name:           "testzone",
												ComputeCluster: "scc01w01-DEV",
											},
										},
									},
								},
								DNSServers: []string{"1.2.3.4"},
								Constraints: apisvsphere.Constraints{
									LoadBalancerConfig: apisvsphere.LoadBalancerConfig{
										Size: "MEDIUM",
										Classes: []apisvsphere.LoadBalancerClass{
											{
												Name:       "default",
												IPPoolName: "lbpool",
											},
											{
												Name:       "private",
												IPPoolName: "lbpool2",
											},
										},
									},
								},
								MachineImages: []apisvsphere.MachineImages{
									{Name: "coreos",
										Versions: []apisvsphere.MachineImageVersion{
											{
												Version: "2191.5.0",
												Path:    "gardener/templates/coreos-2191.5.0",
												GuestID: "coreos64Guest",
											},
										},
									},
								},
							}),
						},
					},
				},
			},
			Shoot: &gardencorev1alpha1.Shoot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "shoot--foo--bar",
					Namespace: namespace,
				},
				Spec: gardencorev1alpha1.ShootSpec{
					Region: "testregion",
					Networking: gardencorev1alpha1.Networking{
						Pods: &cidr,
					},
					Kubernetes: gardencorev1alpha1.Kubernetes{
						Version: "1.14.0",
					},
					Provider: gardencorev1alpha1.Provider{
						ControlPlaneConfig: &gardencorev1alpha1.ProviderConfig{
							runtime.RawExtension{
								Raw: encode(cpConfig),
							},
						},
					},
				},
			},
		}

		cpSecretKey = client.ObjectKey{Namespace: namespace, Name: v1alpha1constants.SecretNameCloudProvider}
		cpSecret    = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      v1alpha1constants.SecretNameCloudProvider,
				Namespace: namespace,
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"vsphereUsername": []byte("admin"),
				"vspherePassword": []byte("super-secret"),
				"nsxtUsername":    []byte("nsxt-admin"),
				"nsxtPassword":    []byte("nsxt-super-secret"),
			},
		}

		checksums = map[string]string{
			v1alpha1constants.SecretNameCloudProvider: "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			vsphere.CloudProviderConfig:               "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606",
			"cloud-controller-manager":                "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
			"cloud-controller-manager-server":         "6dff2a2e6f14444b66d8e4a351c049f7e89ee24ba3eaab95dbec40ba6bdebb52",
			"csi-attacher":                            "2da58ad61c401a2af779a909d22fb42eed93a1524cbfdab974ceedb413fcb914",
			"csi-provisioner":                         "f75b42d40ab501428c383dfb2336cb1fc892bbee1fc1d739675171e4acc4d911",
			vsphere.SecretCsiVsphereConfig:            "5555555555",
			"vsphere-csi-controller":                  "6666666666",
			"csi-vsphere-csi-syncer":                  "7777777777",
		}

		configChartValues = map[string]interface{}{
			"insecureFlag": "1",
			"serverPort":   443,
			"serverName":   "vsphere.host.internal",
			"datacenters":  "scc01-DC",
			"username":     "admin",
			"password":     "super-secret",
			"loadbalancer": map[string]interface{}{
				"size":       "MEDIUM",
				"ipPoolName": "lbpool",
				"classes": []map[string]interface{}{
					{
						"name":       "private",
						"ipPoolName": "lbpool2",
					},
				},
			},
			"nsxt": map[string]interface{}{
				"password":     "nsxt-super-secret",
				"host":         "nsxt.host.internal",
				"insecureFlag": true,
				"username":     "nsxt-admin",
			},
		}

		controlPlaneChartValues = map[string]interface{}{
			"vsphere-cloud-controller-manager": map[string]interface{}{
				"replicas":          1,
				"kubernetesVersion": "1.14.0",
				"clusterName":       "shoot--foo--bar-garden1234",
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
			},
			"nsxt-lb-provider-manager": map[string]interface{}{
				"podAnnotations": map[string]interface{}{
					"checksum/secret-cloud-controller-manager":        "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
					"checksum/secret-cloud-controller-manager-server": "6dff2a2e6f14444b66d8e4a351c049f7e89ee24ba3eaab95dbec40ba6bdebb52",
					"checksum/secret-cloudprovider":                   "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
					"checksum/configmap-cloud-provider-config":        "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606",
				},
				"replicas":          1,
				"clusterName":       "shoot--foo--bar-garden1234",
				"kubernetesVersion": "1.14.0",
			},
			"csi-vsphere": map[string]interface{}{
				"replicas":          1,
				"kubernetesVersion": "1.14.0",
				"serverName":        "vsphere.host.internal",
				"clusterID":         "shoot--foo--bar-garden1234",
				"username":          "admin",
				"password":          "super-secret",
				"serverPort":        443,
				"datacenters":       "scc01-DC",
				"insecureFlag":      "true",
				"podAnnotations": map[string]interface{}{
					"checksum/secret-csi-attacher":              "2da58ad61c401a2af779a909d22fb42eed93a1524cbfdab974ceedb413fcb914",
					"checksum/secret-csi-provisioner":           "f75b42d40ab501428c383dfb2336cb1fc892bbee1fc1d739675171e4acc4d911",
					"checksum/secret-cloudprovider":             "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
					"checksum/secret-csi-vsphere-config-secret": "5555555555",
					"checksum/secret-vsphere-csi-controller":    "6666666666",
					"checksum/secret-vsphere-csi-syncer":        "7777777777",
				},
			},
		}

		controlPlaneShootChartValues = map[string]interface{}{
			"csi-vsphere": map[string]interface{}{
				"serverName":        "vsphere.host.internal",
				"clusterID":         "shoot--foo--bar-garden1234",
				"username":          "admin",
				"password":          "super-secret",
				"serverPort":        443,
				"datacenters":       "scc01-DC",
				"insecureFlag":      "true",
				"kubernetesVersion": "1.14.0",
			},
		}

		logger = log.Log.WithName("test")

		prepareValueProvider = func() genericactuator.ValuesProvider {
			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(context.TODO(), cpSecretKey, &corev1.Secret{}).DoAndReturn(clientGet(cpSecret))

			// Create valuesProvider
			vp := NewValuesProvider(logger, "garden1234")
			err := vp.(inject.Scheme).InjectScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			err = vp.(inject.Client).InjectClient(client)
			Expect(err).NotTo(HaveOccurred())

			return vp
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#GetConfigChartValues", func() {
		It("should return correct config chart values", func() {
			vp := prepareValueProvider()

			// Call GetConfigChartValues method and check the result
			values, err := vp.GetConfigChartValues(context.TODO(), cp, cluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(configChartValues))
		})
	})

	Describe("#GetControlPlaneChartValues", func() {
		It("should return correct control plane chart values", func() {
			vp := prepareValueProvider()

			// Call GetControlPlaneChartValues method and check the result
			values, err := vp.GetControlPlaneChartValues(context.TODO(), cp, cluster, checksums, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(values["csi-vsphere"]).To(Equal(controlPlaneChartValues["csi-vsphere"]))
			Expect(values["vsphere-cloud-controller-manager"]).To(Equal(controlPlaneChartValues["vsphere-cloud-controller-manager"]))
			Expect(values).To(Equal(controlPlaneChartValues))
		})
	})

	Describe("#GetControlPlaneShootChartValues", func() {
		It("should return correct control plane shoot chart values", func() {
			vp := prepareValueProvider()

			// Call GetControlPlaneChartValues method and check the result
			values, err := vp.GetControlPlaneShootChartValues(context.TODO(), cp, cluster, checksums)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(controlPlaneShootChartValues))
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
