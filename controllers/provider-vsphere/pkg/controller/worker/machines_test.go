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

package worker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	api "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere"
	apiv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere/v1alpha1"
	vspherev1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere/v1alpha1"
	. "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/controller/worker"
	"github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/vsphere"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/common"
	"github.com/gardener/gardener-extensions/pkg/controller/worker"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockkubernetes "github.com/gardener/gardener-extensions/pkg/mock/gardener/client/kubernetes"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO martin adapt/fix test

var _ = Describe("Machines", func() {
	var (
		ctrl         *gomock.Controller
		c            *mockclient.MockClient
		chartApplier *mockkubernetes.MockChartApplier
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		c = mockclient.NewMockClient(ctrl)
		chartApplier = mockkubernetes.NewMockChartApplier(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("workerDelegate", func() {
		workerDelegate, _ := NewWorkerDelegate(common.NewClientContext(nil, nil, nil), nil, "", nil, nil)

		Describe("#MachineClassKind", func() {
			It("should return the correct kind of the machine class", func() {
				Expect(workerDelegate.MachineClassKind()).To(Equal("MachineClass"))
			})
		})

		Describe("#MachineClassList", func() {
			It("should return the correct type for the machine class list", func() {
				Expect(workerDelegate.MachineClassList()).To(Equal(&machinev1alpha1.MachineClassList{}))
			})
		})

		Describe("#GenerateMachineDeployments, #DeployMachineClasses", func() {
			var (
				namespace        string
				cloudProfileName string

				host         string
				username     string
				password     string
				nsxtUsername string
				nsxtPassword string
				insecureSSL  bool
				region       string

				machineImageName    string
				machineImageVersion string
				machineImagePath    string

				machineType   string
				networkName   string
				datacenter    string
				resourcePool  string
				datastore     string
				resourcePool2 string
				datastore2    string
				folder        string
				sshKey        string
				userData      = []byte("some-user-data")

				namePool1           string
				minPool1            int
				maxPool1            int
				maxSurgePool1       intstr.IntOrString
				maxUnavailablePool1 intstr.IntOrString

				namePool2           string
				minPool2            int
				maxPool2            int
				maxSurgePool2       intstr.IntOrString
				maxUnavailablePool2 intstr.IntOrString

				zone1 string
				zone2 string

				workerPoolHash1 string
				workerPoolHash2 string

				shootVersionMajorMinor string
				shootVersion           string
				scheme                 *runtime.Scheme
				decoder                runtime.Decoder
				cluster                *extensionscontroller.Cluster
				w                      *extensionsv1alpha.Worker
			)

			BeforeEach(func() {
				namespace = "shoot--foobar--vsphere"
				cloudProfileName = "vsphere"

				region = "testregion"
				host = "vsphere.host.internal"
				username = "myuser"
				password = "mypassword"
				insecureSSL = true
				nsxtUsername = "nsxtuser"
				nsxtPassword = "nsxtpassword"

				machineImageName = "my-os"
				machineImageVersion = "123"
				machineImagePath = "templates/my-template"

				machineType = "mt1"
				datacenter = "my-dc"
				resourcePool = "my-pool"
				datastore = "my-ds"
				resourcePool2 = "my-pool2"
				datastore2 = "my-ds2"
				folder = "my-folder"
				networkName = "mynetwork"
				sshKey = "aaabbbcccddd"
				userData = []byte("some-user-data")

				namePool1 = "pool-1"
				minPool1 = 5
				maxPool1 = 10
				maxSurgePool1 = intstr.FromInt(3)
				maxUnavailablePool1 = intstr.FromInt(2)

				namePool2 = "pool-2"
				minPool2 = 30
				maxPool2 = 45
				maxSurgePool2 = intstr.FromInt(10)
				maxUnavailablePool2 = intstr.FromInt(15)

				zone1 = "testregion-a"
				zone2 = "testregion-b"

				shootVersionMajorMinor = "1.2"
				shootVersion = shootVersionMajorMinor + ".3"

				images := []apiv1alpha1.MachineImages{
					{
						Name: machineImageName,
						Versions: []apiv1alpha1.MachineImageVersion{
							{
								Version: machineImageVersion,
								Path:    machineImagePath,
							},
						},
					},
				}
				cluster = createCluster(cloudProfileName, shootVersion, images)

				w = &extensionsv1alpha.Worker{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
					},
					Spec: extensionsv1alpha.WorkerSpec{
						SecretRef: corev1.SecretReference{
							Name:      "secret",
							Namespace: namespace,
						},
						Region: region,
						InfrastructureProviderStatus: &runtime.RawExtension{
							Raw: encode(&api.InfrastructureStatus{
								Network: networkName,
								VsphereConfig: api.VsphereConfig{
									Folder: folder,
									Region: region,
									ZoneConfigs: map[string]api.ZoneConfig{
										"testregion-a": {
											Datacenter:   datacenter,
											Datastore:    datastore,
											ResourcePool: resourcePool,
										},
										"testregion-b": {
											Datacenter:   datacenter,
											Datastore:    datastore2,
											ResourcePool: resourcePool2,
										},
									},
								},
							}),
						},
						Pools: []extensionsv1alpha.WorkerPool{
							{
								Name:           namePool1,
								Minimum:        minPool1,
								Maximum:        maxPool1,
								MaxSurge:       maxSurgePool1,
								MaxUnavailable: maxUnavailablePool1,
								MachineType:    machineType,
								MachineImage: extensionsv1alpha.MachineImage{
									Name:    machineImageName,
									Version: machineImageVersion,
								},
								UserData: userData,
								Zones: []string{
									zone1,
									zone2,
								},
							},
							{
								Name:           namePool2,
								Minimum:        minPool2,
								Maximum:        maxPool2,
								MaxSurge:       maxSurgePool2,
								MaxUnavailable: maxUnavailablePool2,
								MachineType:    machineType,
								MachineImage: extensionsv1alpha.MachineImage{
									Name:    machineImageName,
									Version: machineImageVersion,
								},
								UserData: userData,
								Zones: []string{
									zone1,
									zone2,
								},
							},
						},
						SSHPublicKey: []byte(sshKey),
					},
				}

				scheme = runtime.NewScheme()
				_ = api.AddToScheme(scheme)
				_ = apiv1alpha1.AddToScheme(scheme)
				decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()

				workerPoolHash1, _ = worker.WorkerPoolHash(w.Spec.Pools[0], cluster)
				workerPoolHash2, _ = worker.WorkerPoolHash(w.Spec.Pools[1], cluster)

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)
			})

			It("should return the expected machine deployments", func() {
				expectGetSecretCallToWork(c, username, password, nsxtUsername, nsxtPassword)

				// Test workerDelegate.DeployMachineClasses()
				defaultMachineClass := map[string]interface{}{
					"region":       region,
					"resourcePool": resourcePool,
					"datacenter":   datacenter,
					"datastore":    datastore,
					"folder":       folder,
					"network":      networkName,
					"memory":       4096,
					"numCpus":      2,
					"systemDisk": map[string]interface{}{
						"size": 20,
					},
					"templateVM": machineImagePath,
					"sshKeys":    []string{sshKey},
					"tags": map[string]string{
						fmt.Sprintf("kubernetes.io/cluster/%s", namespace): "1",
						"kubernetes.io/role/node":                          "1",
					},
					"secret": map[string]interface{}{
						"cloudConfig": string(userData),
					},
				}

				machineClassNamePool1Zone1 := fmt.Sprintf("%s-%s-z1", namespace, namePool1)
				machineClassNamePool1Zone2 := fmt.Sprintf("%s-%s-z2", namespace, namePool1)
				machineClassNamePool2Zone1 := fmt.Sprintf("%s-%s-z1", namespace, namePool2)
				machineClassNamePool2Zone2 := fmt.Sprintf("%s-%s-z2", namespace, namePool2)

				machineClassPool1Zone1 := prepareMachineClass(defaultMachineClass, machineClassNamePool1Zone1, resourcePool, datastore, workerPoolHash1, host, username, password, insecureSSL)
				machineClassPool1Zone2 := prepareMachineClass(defaultMachineClass, machineClassNamePool1Zone2, resourcePool2, datastore2, workerPoolHash1, host, username, password, insecureSSL)
				machineClassPool2Zone1 := prepareMachineClass(defaultMachineClass, machineClassNamePool2Zone1, resourcePool, datastore, workerPoolHash2, host, username, password, insecureSSL)
				machineClassPool2Zone2 := prepareMachineClass(defaultMachineClass, machineClassNamePool2Zone2, resourcePool2, datastore2, workerPoolHash2, host, username, password, insecureSSL)

				machineClassWithHashPool1Zone1 := machineClassPool1Zone1["name"].(string)
				machineClassWithHashPool1Zone2 := machineClassPool1Zone2["name"].(string)
				machineClassWithHashPool2Zone1 := machineClassPool2Zone1["name"].(string)
				machineClassWithHashPool2Zone2 := machineClassPool2Zone2["name"].(string)

				chartApplier.
					EXPECT().
					ApplyChart(
						context.TODO(),
						filepath.Join(vsphere.InternalChartsPath, "machineclass"),
						namespace,
						"machineclass",
						map[string]interface{}{"machineClasses": []map[string]interface{}{
							machineClassPool1Zone1,
							machineClassPool1Zone2,
							machineClassPool2Zone1,
							machineClassPool2Zone2,
						}},
						nil,
					).
					Return(nil)

				err := workerDelegate.DeployMachineClasses(context.TODO())
				Expect(err).NotTo(HaveOccurred())

				// Test workerDelegate.GetMachineImages()
				machineImages, err := workerDelegate.GetMachineImages(context.TODO())
				Expect(machineImages).To(Equal(&vspherev1alpha1.WorkerStatus{
					TypeMeta: metav1.TypeMeta{
						APIVersion: vspherev1alpha1.SchemeGroupVersion.String(),
						Kind:       "WorkerStatus",
					},
					MachineImages: []vspherev1alpha1.MachineImage{
						{
							Name:    machineImageName,
							Version: machineImageVersion,
							Path:    machineImagePath,
						},
					},
				}))
				Expect(err).NotTo(HaveOccurred())

				// Test workerDelegate.GenerateMachineDeployments()
				machineDeployments := worker.MachineDeployments{
					{
						Name:           machineClassNamePool1Zone1,
						ClassName:      machineClassWithHashPool1Zone1,
						SecretName:     machineClassWithHashPool1Zone1,
						Minimum:        minPool1,
						Maximum:        maxPool1,
						MaxSurge:       maxSurgePool1,
						MaxUnavailable: maxUnavailablePool1,
					},
					{
						Name:           machineClassNamePool1Zone2,
						ClassName:      machineClassWithHashPool1Zone2,
						SecretName:     machineClassWithHashPool1Zone2,
						Minimum:        minPool1,
						Maximum:        maxPool1,
						MaxSurge:       maxSurgePool1,
						MaxUnavailable: maxUnavailablePool1,
					},
					{
						Name:           machineClassNamePool2Zone1,
						ClassName:      machineClassWithHashPool2Zone1,
						SecretName:     machineClassWithHashPool2Zone1,
						Minimum:        minPool2,
						Maximum:        maxPool2,
						MaxSurge:       maxSurgePool2,
						MaxUnavailable: maxUnavailablePool2,
					},
					{
						Name:           machineClassNamePool2Zone2,
						ClassName:      machineClassWithHashPool2Zone2,
						SecretName:     machineClassWithHashPool2Zone2,
						Minimum:        minPool2,
						Maximum:        maxPool2,
						MaxSurge:       maxSurgePool2,
						MaxUnavailable: maxUnavailablePool2,
					},
				}

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(machineDeployments))
			})

			It("should fail because the secret cannot be read", func() {
				c.EXPECT().
					Get(context.TODO(), gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
					Return(fmt.Errorf("error"))

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the version is invalid", func() {
				expectGetSecretCallToWork(c, username, password, nsxtUsername, nsxtPassword)

				cluster.Shoot.Spec.Kubernetes.Version = "invalid"
				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the infrastructure status cannot be decoded", func() {
				expectGetSecretCallToWork(c, username, password, nsxtUsername, nsxtPassword)

				w.Spec.InfrastructureProviderStatus = &runtime.RawExtension{Raw: []byte(`invalid`)}

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the machine image cannot be found", func() {
				expectGetSecretCallToWork(c, username, password, nsxtUsername, nsxtPassword)

				invalidImages := []apiv1alpha1.MachineImages{
					{
						Name: "xxname",
						Versions: []apiv1alpha1.MachineImageVersion{
							{
								Version: "xxversion",
								Path:    "xxpath",
							},
						},
					},
				}
				clusterWithoutImages := createCluster(cloudProfileName, shootVersion, invalidImages)

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, clusterWithoutImages)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})
})

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}

func expectGetSecretCallToWork(c *mockclient.MockClient, username, password string, nsxtUsername, nsxtPassword string) {
	c.EXPECT().
		Get(context.TODO(), gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
		DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
			secret.Data = map[string][]byte{
				vsphere.Username:     []byte(username),
				vsphere.Password:     []byte(password),
				vsphere.NSXTUsername: []byte(nsxtUsername),
				vsphere.NSXTPassword: []byte(nsxtPassword),
			}
			return nil
		})
}

func createCluster(cloudProfileName, shootVersion string, images []apiv1alpha1.MachineImages) *extensionscontroller.Cluster {
	cloudProfileConfig := &apiv1alpha1.CloudProfileConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiv1alpha1.SchemeGroupVersion.String(),
			Kind:       "CloudProfileConfig",
		},
		NamePrefix:                    "nameprefix",
		DefaultClassStoragePolicyName: "mypolicy",
		DNSServers:                    []string{"1.2.3.4"},
		Regions: []apiv1alpha1.RegionSpec{
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
				Zones: []apiv1alpha1.ZoneSpec{
					{
						Name:           "testregion-a",
						ComputeCluster: "scc01w01-DEV-A",
					},
					{
						Name:           "testregion-b",
						ComputeCluster: "scc01w01-DEV-B",
					},
				},
			},
		},
		Constraints: apiv1alpha1.Constraints{
			LoadBalancerConfig: apiv1alpha1.LoadBalancerConfig{
				Size: "SMALL",
				Classes: []apiv1alpha1.LoadBalancerClass{
					{
						Name:       "default",
						IPPoolName: "lbpool",
					},
				},
			},
		},
		MachineImages: images,
	}
	cloudProfileConfigJSON, _ := json.Marshal(cloudProfileConfig)
	cluster := &extensionscontroller.Cluster{
		CloudProfile: &gardencorev1alpha1.CloudProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name: cloudProfileName,
			},
			Spec: gardencorev1alpha1.CloudProfileSpec{
				ProviderConfig: &gardencorev1alpha1.ProviderConfig{
					RawExtension: runtime.RawExtension{
						Raw: cloudProfileConfigJSON,
					},
				},
				Regions: []gardencorev1alpha1.Region{
					{
						Name: "testregion",
						Zones: []gardencorev1alpha1.AvailabilityZone{
							{Name: "testregion-a"},
							{Name: "testregion-b"},
						},
					},
				},
				MachineTypes: []gardencorev1alpha1.MachineType{
					{
						Name:   "mt1",
						Memory: resource.MustParse("4096Mi"),
						CPU:    resource.MustParse("2"),
					},
				},
			},
		},
		Shoot: &gardencorev1alpha1.Shoot{
			Spec: gardencorev1alpha1.ShootSpec{
				Region: "testregion",
				Kubernetes: gardencorev1alpha1.Kubernetes{
					Version: shootVersion,
				},
			},
		},
	}

	return cluster
}

func prepareMachineClass(defaultMachineClass map[string]interface{}, machineClassName, resourcePool, datastore, workerPoolHash, host, username, password string, insecureSSL bool) map[string]interface{} {
	out := make(map[string]interface{}, len(defaultMachineClass)+10)

	for k, v := range defaultMachineClass {
		out[k] = v
	}

	out["resourcePool"] = resourcePool
	out["datastore"] = datastore
	out["name"] = fmt.Sprintf("%s-%s", machineClassName, workerPoolHash)
	out["secret"].(map[string]interface{})[vsphere.Host] = host
	out["secret"].(map[string]interface{})[vsphere.Username] = username
	out["secret"].(map[string]interface{})[vsphere.Password] = password
	out["secret"].(map[string]interface{})[vsphere.InsecureSSL] = strconv.FormatBool(insecureSSL)

	return out
}
