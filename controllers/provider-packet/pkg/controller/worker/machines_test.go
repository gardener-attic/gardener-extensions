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

	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/apis/config"
	apispacket "github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/apis/packet"
	packetv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/apis/packet/v1alpha1"
	. "github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/controller/worker"
	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/packet"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/worker"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockkubernetes "github.com/gardener/gardener-extensions/pkg/mock/gardener/client/kubernetes"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
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
				Expect(workerDelegate.MachineClassKind()).To(Equal("PacketMachineClass"))
			})
		})

		Describe("#MachineClassList", func() {
			It("should return the correct type for the machine class list", func() {
				Expect(workerDelegate.MachineClassList()).To(Equal(&machinev1alpha1.PacketMachineClassList{}))
			})
		})

		Describe("#GenerateMachineDeployments, #DeployMachineClasses", func() {
			var (
				namespace string

				packetAPIToken  string
				packetProjectID string
				region          string

				machineImageName    string
				machineImageVersion string
				machineImage        string

				machineType string
				sshKeyID    string
				userData    = []byte("some-user-data")

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

				zone1 = region + "a"
				zone2 = region + "b"

				workerPoolHash1 string
				workerPoolHash2 string

				shootVersionMajorMinor string
				shootVersion           string
				machineImages          []config.MachineImage
				scheme                 *runtime.Scheme
				decoder                runtime.Decoder
				cluster                *extensionscontroller.Cluster
				w                      *extensionsv1alpha.Worker
			)

			BeforeEach(func() {
				namespace = "shoot--foobar--packet"

				region = "eu-west-1"
				packetAPIToken = "api-token"
				packetProjectID = "project-id"

				machineImageName = "my-os"
				machineImageVersion = "123"
				machineImage = "uuid"

				machineType = "large"
				sshKeyID = "1-2-3-4"
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

				zone1 = region + "a"
				zone2 = region + "b"

				shootVersionMajorMinor = "1.2"
				shootVersion = shootVersionMajorMinor + ".3"

				machineImages = []config.MachineImage{
					{
						Name:    machineImageName,
						Version: machineImageVersion,
						ID:      machineImage,
					},
				}

				cluster = &extensionscontroller.Cluster{
					Shoot: &gardencorev1beta1.Shoot{
						Spec: gardencorev1beta1.ShootSpec{
							Kubernetes: gardencorev1beta1.Kubernetes{
								Version: shootVersion,
							},
						},
					},
				}

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
							Raw: encode(&apispacket.InfrastructureStatus{
								SSHKeyID: sshKeyID,
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
					},
				}

				scheme = runtime.NewScheme()
				_ = apispacket.AddToScheme(scheme)
				_ = packetv1alpha1.AddToScheme(scheme)
				decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()

				workerPoolHash1, _ = worker.WorkerPoolHash(w.Spec.Pools[0], cluster)
				workerPoolHash2, _ = worker.WorkerPoolHash(w.Spec.Pools[1], cluster)

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImages, chartApplier, "", w, cluster)
			})

			It("should return the expected machine deployments", func() {
				expectGetSecretCallToWork(c, packetAPIToken, packetProjectID)

				// Test workerDelegate.DeployMachineClasses()
				var (
					defaultMachineClass = map[string]interface{}{
						"OS":           machineImage,
						"projectID":    packetProjectID,
						"billingCycle": "hourly",
						"machineType":  machineType,
						"facility": []string{
							zone1,
							zone2,
						},
						"sshKeys": []string{sshKeyID},
						"tags": []string{
							fmt.Sprintf("kubernetes.io/cluster/%s", namespace),
							"kubernetes.io/role/node",
						},
						"secret": map[string]interface{}{
							"cloudConfig": string(userData),
						},
					}

					machineClassPool1 = copyMachineClass(defaultMachineClass)
					machineClassPool2 = copyMachineClass(defaultMachineClass)

					machineClassNamePool1 = fmt.Sprintf("%s-%s", namespace, namePool1)
					machineClassNamePool2 = fmt.Sprintf("%s-%s", namespace, namePool2)

					machineClassWithHashPool1 = fmt.Sprintf("%s-%s", machineClassNamePool1, workerPoolHash1)
					machineClassWithHashPool2 = fmt.Sprintf("%s-%s", machineClassNamePool2, workerPoolHash2)
				)

				addNameAndSecretToMachineClass(machineClassPool1, packetAPIToken, machineClassWithHashPool1)
				addNameAndSecretToMachineClass(machineClassPool2, packetAPIToken, machineClassWithHashPool2)

				chartApplier.
					EXPECT().
					ApplyChart(
						context.TODO(),
						filepath.Join(packet.InternalChartsPath, "machineclass"),
						namespace,
						"machineclass",
						map[string]interface{}{"machineClasses": []map[string]interface{}{
							machineClassPool1,
							machineClassPool2,
						}},
						nil,
					).
					Return(nil)

				err := workerDelegate.DeployMachineClasses(context.TODO())
				Expect(err).NotTo(HaveOccurred())

				// Test workerDelegate.GetMachineImages()
				machineImages, err := workerDelegate.GetMachineImages(context.TODO())
				Expect(machineImages).To(Equal(&packetv1alpha1.WorkerStatus{
					TypeMeta: metav1.TypeMeta{
						APIVersion: packetv1alpha1.SchemeGroupVersion.String(),
						Kind:       "WorkerStatus",
					},
					MachineImages: []packetv1alpha1.MachineImage{
						{
							Name:    machineImageName,
							Version: machineImageVersion,
							ID:      machineImage,
						},
					},
				}))
				Expect(err).NotTo(HaveOccurred())

				// Test workerDelegate.GenerateMachineDeployments()
				machineDeployments := worker.MachineDeployments{
					{
						Name:           machineClassNamePool1,
						ClassName:      machineClassWithHashPool1,
						SecretName:     machineClassWithHashPool1,
						Minimum:        minPool1,
						Maximum:        maxPool1,
						MaxSurge:       maxSurgePool1,
						MaxUnavailable: maxUnavailablePool1,
					},
					{
						Name:           machineClassNamePool2,
						ClassName:      machineClassWithHashPool2,
						SecretName:     machineClassWithHashPool2,
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
				expectGetSecretCallToWork(c, packetAPIToken, packetProjectID)

				cluster.Shoot.Spec.Kubernetes.Version = "invalid"
				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImages, chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the infrastructure status cannot be decoded", func() {
				expectGetSecretCallToWork(c, packetAPIToken, packetProjectID)

				w.Spec.InfrastructureProviderStatus = &runtime.RawExtension{Raw: []byte(`invalid`)}

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImages, chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the machine image cannot be found", func() {
				expectGetSecretCallToWork(c, packetAPIToken, packetProjectID)

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, nil, chartApplier, "", w, cluster)

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

func expectGetSecretCallToWork(c *mockclient.MockClient, packetAPIToken, packetProjectID string) {
	c.EXPECT().
		Get(context.TODO(), gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
		DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
			secret.Data = map[string][]byte{
				packet.APIToken:  []byte(packetAPIToken),
				packet.ProjectID: []byte(packetProjectID),
			}
			return nil
		})
}

func copyMachineClass(def map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(def))

	for k, v := range def {
		out[k] = v
	}

	return out
}

func addNameAndSecretToMachineClass(class map[string]interface{}, packetAPIToken, name string) {
	class["name"] = name
	class["secret"].(map[string]interface{})[packet.APIToken] = packetAPIToken
}
