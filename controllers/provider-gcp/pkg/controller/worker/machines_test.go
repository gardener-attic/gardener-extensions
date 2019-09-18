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

package worker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/config"
	apisgcp "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp"
	gcpv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp/v1alpha1"
	. "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/controller/worker"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/gcp"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/worker"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockkubernetes "github.com/gardener/gardener-extensions/pkg/mock/gardener/client/kubernetes"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
		workerDelegate := NewWorkerDelegate(nil, nil, nil, nil, nil, "", nil, nil)

		Describe("#MachineClassKind", func() {
			It("should return the correct kind of the machine class", func() {
				Expect(workerDelegate.MachineClassKind()).To(Equal("GCPMachineClass"))
			})
		})

		Describe("#MachineClassList", func() {
			It("should return the correct type for the machine class list", func() {
				Expect(workerDelegate.MachineClassList()).To(Equal(&machinev1alpha1.GCPMachineClassList{}))
			})
		})

		Describe("#GenerateMachineDeployments, #DeployMachineClasses", func() {
			var (
				name      string
				namespace string

				serviceAccountJSON string
				region             string

				machineImageName    string
				machineImageVersion string
				machineImage        string

				serviceAccountEmail string
				machineType         string
				userData            = []byte("some-user-data")
				subnetName          string

				volumeType string
				volumeSize int

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

				shootVersionMajorMinor string
				shootVersion           string
				machineImages          []config.MachineImage
				scheme                 *runtime.Scheme
				decoder                runtime.Decoder
				cluster                *extensionscontroller.Cluster
				w                      *extensionsv1alpha1.Worker
			)

			BeforeEach(func() {
				name = "my-shoot"
				namespace = "shoot--foobar--gcp"

				region = "eu-west-1"
				serviceAccountJSON = "some-json-doc"

				machineImageName = "my-os"
				machineImageVersion = "123"
				machineImage = "path/to/project/machine/image"

				serviceAccountEmail = "service@account.com"
				machineType = "large"
				userData = []byte("some-user-data")
				subnetName = "subnet-nodes"

				volumeType = "normal"
				volumeSize = 20

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

				zone1 = region + "a"
				zone2 = region + "b"

				shootVersionMajorMinor = "1.2"
				shootVersion = shootVersionMajorMinor + ".3"

				machineImages = []config.MachineImage{
					{
						Name:    machineImageName,
						Version: machineImageVersion,
						Image:   machineImage,
					},
				}

				cluster = &extensionscontroller.Cluster{
					CoreShoot: &gardencorev1alpha1.Shoot{
						Spec: gardencorev1alpha1.ShootSpec{
							Kubernetes: gardencorev1alpha1.Kubernetes{
								Version: shootVersion,
							},
						},
					},
				}

				w = &extensionsv1alpha1.Worker{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
					Spec: extensionsv1alpha1.WorkerSpec{
						SecretRef: corev1.SecretReference{
							Name:      "secret",
							Namespace: namespace,
						},
						Region: region,
						InfrastructureProviderStatus: &runtime.RawExtension{
							Raw: encode(&apisgcp.InfrastructureStatus{
								ServiceAccountEmail: serviceAccountEmail,
								Networks: apisgcp.NetworkStatus{
									Subnets: []apisgcp.Subnet{
										{
											Name:    subnetName,
											Purpose: apisgcp.PurposeNodes,
										},
									},
								},
							}),
						},
						Pools: []extensionsv1alpha1.WorkerPool{
							{
								Name:           namePool1,
								Minimum:        minPool1,
								Maximum:        maxPool1,
								MaxSurge:       maxSurgePool1,
								MaxUnavailable: maxUnavailablePool1,
								MachineType:    machineType,
								MachineImage: extensionsv1alpha1.MachineImage{
									Name:    machineImageName,
									Version: machineImageVersion,
								},
								UserData: userData,
								Volume: &extensionsv1alpha1.Volume{
									Type: volumeType,
									Size: fmt.Sprintf("%dGi", volumeSize),
								},
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
								MachineImage: extensionsv1alpha1.MachineImage{
									Name:    machineImageName,
									Version: machineImageVersion,
								},
								UserData: userData,
								Volume: &extensionsv1alpha1.Volume{
									Type: volumeType,
									Size: fmt.Sprintf("%dGi", volumeSize),
								},
								Zones: []string{
									zone1,
									zone2,
								},
							},
						},
					},
				}

				scheme = runtime.NewScheme()
				_ = apisgcp.AddToScheme(scheme)
				_ = gcpv1alpha1.AddToScheme(scheme)
				decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImages, chartApplier, "", w, cluster)
			})

			It("should return the expected machine deployments", func() {
				expectGetSecretCallToWork(c, serviceAccountJSON)

				// Test workerDelegate.DeployMachineClasses()
				var (
					defaultMachineClass = map[string]interface{}{
						"region":             region,
						"canIpForward":       true,
						"deletionProtection": false,
						"description":        fmt.Sprintf("Machine of Shoot %s created by machine-controller-manager.", name),
						"disks": []map[string]interface{}{
							{
								"autoDelete": true,
								"boot":       true,
								"sizeGb":     volumeSize,
								"type":       volumeType,
								"image":      machineImage,
								"labels": map[string]interface{}{
									"name": name,
								},
							},
						},
						"labels": map[string]interface{}{
							"name": name,
						},
						"machineType": machineType,
						"networkInterfaces": []map[string]interface{}{
							{
								"subnetwork": subnetName,
							},
						},
						"scheduling": map[string]interface{}{
							"automaticRestart":  true,
							"onHostMaintenance": "MIGRATE",
							"preemptible":       false,
						},
						"secret": map[string]interface{}{
							"cloudConfig": string(userData),
						},
						"serviceAccounts": []map[string]interface{}{
							{
								"email": serviceAccountEmail,
								"scopes": []string{
									"https://www.googleapis.com/auth/compute",
								},
							},
						},
						"tags": []string{
							namespace,
							fmt.Sprintf("kubernetes-io-cluster-%s", namespace),
							"kubernetes-io-role-node",
						},
					}

					machineClassPool1Zone1 = useDefaultMachineClass(defaultMachineClass, "zone", zone1)
					machineClassPool1Zone2 = useDefaultMachineClass(defaultMachineClass, "zone", zone2)
					machineClassPool2Zone1 = useDefaultMachineClass(defaultMachineClass, "zone", zone1)
					machineClassPool2Zone2 = useDefaultMachineClass(defaultMachineClass, "zone", zone2)

					machineClassNamePool1Zone1 = fmt.Sprintf("%s-%s-z1", namespace, namePool1)
					machineClassNamePool1Zone2 = fmt.Sprintf("%s-%s-z2", namespace, namePool1)
					machineClassNamePool2Zone1 = fmt.Sprintf("%s-%s-z1", namespace, namePool2)
					machineClassNamePool2Zone2 = fmt.Sprintf("%s-%s-z2", namespace, namePool2)

					machineClassHashPool1Zone1 = worker.MachineClassHash(machineClassPool1Zone1, shootVersionMajorMinor)
					machineClassHashPool1Zone2 = worker.MachineClassHash(machineClassPool1Zone2, shootVersionMajorMinor)
					machineClassHashPool2Zone1 = worker.MachineClassHash(machineClassPool2Zone1, shootVersionMajorMinor)
					machineClassHashPool2Zone2 = worker.MachineClassHash(machineClassPool2Zone2, shootVersionMajorMinor)

					machineClassWithHashPool1Zone1 = fmt.Sprintf("%s-%s", machineClassNamePool1Zone1, machineClassHashPool1Zone1)
					machineClassWithHashPool1Zone2 = fmt.Sprintf("%s-%s", machineClassNamePool1Zone2, machineClassHashPool1Zone2)
					machineClassWithHashPool2Zone1 = fmt.Sprintf("%s-%s", machineClassNamePool2Zone1, machineClassHashPool2Zone1)
					machineClassWithHashPool2Zone2 = fmt.Sprintf("%s-%s", machineClassNamePool2Zone2, machineClassHashPool2Zone2)
				)

				addNameAndSecretToMachineClass(machineClassPool1Zone1, serviceAccountJSON, machineClassWithHashPool1Zone1)
				addNameAndSecretToMachineClass(machineClassPool1Zone2, serviceAccountJSON, machineClassWithHashPool1Zone2)
				addNameAndSecretToMachineClass(machineClassPool2Zone1, serviceAccountJSON, machineClassWithHashPool2Zone1)
				addNameAndSecretToMachineClass(machineClassPool2Zone2, serviceAccountJSON, machineClassWithHashPool2Zone2)

				chartApplier.
					EXPECT().
					ApplyChart(
						context.TODO(),
						filepath.Join(gcp.InternalChartsPath, "machineclass"),
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
				Expect(machineImages).To(Equal(&gcpv1alpha1.WorkerStatus{
					TypeMeta: metav1.TypeMeta{
						APIVersion: gcpv1alpha1.SchemeGroupVersion.String(),
						Kind:       "WorkerStatus",
					},
					MachineImages: []gcpv1alpha1.MachineImage{
						{
							Name:    machineImageName,
							Version: machineImageVersion,
							Image:   machineImage,
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
						Minimum:        worker.DistributeOverZones(0, minPool1, 2),
						Maximum:        worker.DistributeOverZones(0, maxPool1, 2),
						MaxSurge:       worker.DistributePositiveIntOrPercent(0, maxSurgePool1, 2, maxPool1),
						MaxUnavailable: worker.DistributePositiveIntOrPercent(0, maxUnavailablePool1, 2, minPool1),
					},
					{
						Name:           machineClassNamePool1Zone2,
						ClassName:      machineClassWithHashPool1Zone2,
						SecretName:     machineClassWithHashPool1Zone2,
						Minimum:        worker.DistributeOverZones(1, minPool1, 2),
						Maximum:        worker.DistributeOverZones(1, maxPool1, 2),
						MaxSurge:       worker.DistributePositiveIntOrPercent(1, maxSurgePool1, 2, maxPool1),
						MaxUnavailable: worker.DistributePositiveIntOrPercent(1, maxUnavailablePool1, 2, minPool1),
					},
					{
						Name:           machineClassNamePool2Zone1,
						ClassName:      machineClassWithHashPool2Zone1,
						SecretName:     machineClassWithHashPool2Zone1,
						Minimum:        worker.DistributeOverZones(0, minPool2, 2),
						Maximum:        worker.DistributeOverZones(0, maxPool2, 2),
						MaxSurge:       worker.DistributePositiveIntOrPercent(0, maxSurgePool2, 2, maxPool2),
						MaxUnavailable: worker.DistributePositiveIntOrPercent(0, maxUnavailablePool2, 2, minPool2),
					},
					{
						Name:           machineClassNamePool2Zone2,
						ClassName:      machineClassWithHashPool2Zone2,
						SecretName:     machineClassWithHashPool2Zone2,
						Minimum:        worker.DistributeOverZones(1, minPool2, 2),
						Maximum:        worker.DistributeOverZones(1, maxPool2, 2),
						MaxSurge:       worker.DistributePositiveIntOrPercent(1, maxSurgePool2, 2, maxPool2),
						MaxUnavailable: worker.DistributePositiveIntOrPercent(1, maxUnavailablePool2, 2, minPool2),
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
				expectGetSecretCallToWork(c, serviceAccountJSON)

				cluster.CoreShoot.Spec.Kubernetes.Version = "invalid"
				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImages, chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the infrastructure status cannot be decoded", func() {
				expectGetSecretCallToWork(c, serviceAccountJSON)

				w.Spec.InfrastructureProviderStatus = &runtime.RawExtension{}

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImages, chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the nodes subnet cannot be found", func() {
				expectGetSecretCallToWork(c, serviceAccountJSON)

				w.Spec.InfrastructureProviderStatus = &runtime.RawExtension{
					Raw: encode(&apisgcp.InfrastructureStatus{}),
				}

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImages, chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the machine image cannot be found", func() {
				expectGetSecretCallToWork(c, serviceAccountJSON)

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, nil, chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the volume size cannot be decoded", func() {
				expectGetSecretCallToWork(c, serviceAccountJSON)

				w.Spec.Pools[0].Volume.Size = "not-decodeable"

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImages, chartApplier, "", w, cluster)

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

func expectGetSecretCallToWork(c *mockclient.MockClient, serviceAccountJSON string) {
	c.EXPECT().
		Get(context.TODO(), gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
		DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
			secret.Data = map[string][]byte{
				gcp.ServiceAccountJSONField: []byte(serviceAccountJSON),
			}
			return nil
		})
}

func useDefaultMachineClass(def map[string]interface{}, key string, value interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(def)+1)

	for k, v := range def {
		out[k] = v
	}

	out[key] = value
	return out
}

func addNameAndSecretToMachineClass(class map[string]interface{}, serviceAccountJSON, name string) {
	class["name"] = name
	class["secret"].(map[string]interface{})[gcp.ServiceAccountJSONMCM] = serviceAccountJSON
}
