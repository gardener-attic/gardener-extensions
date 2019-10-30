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

	"github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/config"
	apisopenstack "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack"
	openstackv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack/v1alpha1"
	. "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/controller/worker"
	"github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/openstack"
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
				Expect(workerDelegate.MachineClassKind()).To(Equal("OpenStackMachineClass"))
			})
		})

		Describe("#MachineClassList", func() {
			It("should return the correct type for the machine class list", func() {
				Expect(workerDelegate.MachineClassList()).To(Equal(&machinev1alpha1.OpenStackMachineClassList{}))
			})
		})

		Describe("#GenerateMachineDeployments, #DeployMachineClasses", func() {
			var (
				namespace        string
				cloudProfileName string

				openstackAuthURL    string
				openstackDomainName string
				openstackTenantName string
				openstackUserName   string
				openstackPassword   string
				region              string

				machineImageName    string
				machineImageVersion string
				machineImage        string

				keyName           string
				machineType       string
				userData          []byte
				networkID         string
				podCIDR           string
				securityGroupName string

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

				shootVersionMajorMinor             string
				shootVersion                       string
				machineImageToCloudProfilesMapping []config.MachineImage
				scheme                             *runtime.Scheme
				decoder                            runtime.Decoder
				cloudProfileConfig                 *openstackv1alpha1.CloudProfileConfig
				cloudProfileConfigJSON             []byte
				cluster                            *extensionscontroller.Cluster
				w                                  *extensionsv1alpha1.Worker
			)

			BeforeEach(func() {
				namespace = "shoot--foobar--openstack"
				cloudProfileName = "openstack"

				region = "eu-de-1"
				openstackAuthURL = "auth-url"
				openstackDomainName = "domain-name"
				openstackTenantName = "tenant-name"
				openstackUserName = "user-name"
				openstackPassword = "password"

				machineImageName = "my-os"
				machineImageVersion = "123"
				machineImage = "my-image-in-glance"

				keyName = "key-name"
				machineType = "large"
				userData = []byte("some-user-data")
				networkID = "network-id"
				podCIDR = "1.2.3.4/5"
				securityGroupName = "nodes-sec-group"

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

				machineImageToCloudProfilesMapping = []config.MachineImage{
					{
						Name:    machineImageName,
						Version: machineImageVersion,
						CloudProfiles: []config.CloudProfileMapping{
							{
								Name:  cloudProfileName,
								Image: machineImage,
							},
						},
					},
				}

				cloudProfileConfig = &openstackv1alpha1.CloudProfileConfig{
					KeyStoneURL: openstackAuthURL,
				}
				cloudProfileConfigJSON, _ = json.Marshal(cloudProfileConfig)
				cluster = &extensionscontroller.Cluster{
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
						},
					},
					Shoot: &gardencorev1alpha1.Shoot{
						Spec: gardencorev1alpha1.ShootSpec{
							Networking: gardencorev1alpha1.Networking{
								Pods: &podCIDR,
							},
							Kubernetes: gardencorev1alpha1.Kubernetes{
								Version: shootVersion,
							},
						},
					},
				}

				w = &extensionsv1alpha1.Worker{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
					},
					Spec: extensionsv1alpha1.WorkerSpec{
						SecretRef: corev1.SecretReference{
							Name:      "secret",
							Namespace: namespace,
						},
						Region: region,
						InfrastructureProviderStatus: &runtime.RawExtension{
							Raw: encode(&apisopenstack.InfrastructureStatus{
								SecurityGroups: []apisopenstack.SecurityGroup{
									{
										Purpose: apisopenstack.PurposeNodes,
										Name:    securityGroupName,
									},
								},
								Node: apisopenstack.NodeStatus{
									KeyName: keyName,
								},
								Networks: apisopenstack.NetworkStatus{
									ID: networkID,
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
								Zones: []string{
									zone1,
									zone2,
								},
							},
						},
					},
				}

				scheme = runtime.NewScheme()
				_ = apisopenstack.AddToScheme(scheme)
				_ = openstackv1alpha1.AddToScheme(scheme)
				decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImageToCloudProfilesMapping, chartApplier, "", w, cluster)
			})

			It("should return the expected machine deployments", func() {
				expectGetSecretCallToWork(c, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword)

				// Test workerDelegate.DeployMachineClasses()
				var (
					defaultMachineClass = map[string]interface{}{
						"region":         region,
						"machineType":    machineType,
						"keyName":        keyName,
						"imageName":      machineImage,
						"networkID":      networkID,
						"podNetworkCidr": podCIDR,
						"securityGroups": []string{securityGroupName},
						"tags": map[string]string{
							fmt.Sprintf("kubernetes.io-cluster-%s", namespace): "1",
							"kubernetes.io-role-node":                          "1",
						},
						"secret": map[string]interface{}{
							"cloudConfig": string(userData),
						},
					}

					machineClassPool1Zone1 = useDefaultMachineClass(defaultMachineClass, "availabilityZone", zone1)
					machineClassPool1Zone2 = useDefaultMachineClass(defaultMachineClass, "availabilityZone", zone2)
					machineClassPool2Zone1 = useDefaultMachineClass(defaultMachineClass, "availabilityZone", zone1)
					machineClassPool2Zone2 = useDefaultMachineClass(defaultMachineClass, "availabilityZone", zone2)

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

				addNameAndSecretToMachineClass(machineClassPool1Zone1, openstackAuthURL, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword, machineClassWithHashPool1Zone1)
				addNameAndSecretToMachineClass(machineClassPool1Zone2, openstackAuthURL, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword, machineClassWithHashPool1Zone2)
				addNameAndSecretToMachineClass(machineClassPool2Zone1, openstackAuthURL, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword, machineClassWithHashPool2Zone1)
				addNameAndSecretToMachineClass(machineClassPool2Zone2, openstackAuthURL, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword, machineClassWithHashPool2Zone2)

				chartApplier.
					EXPECT().
					ApplyChart(
						context.TODO(),
						filepath.Join(openstack.InternalChartsPath, "machineclass"),
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
				Expect(machineImages).To(Equal(&openstackv1alpha1.WorkerStatus{
					TypeMeta: metav1.TypeMeta{
						APIVersion: openstackv1alpha1.SchemeGroupVersion.String(),
						Kind:       "WorkerStatus",
					},
					MachineImages: []openstackv1alpha1.MachineImage{
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
				expectGetSecretCallToWork(c, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword)

				cluster.Shoot.Spec.Kubernetes.Version = "invalid"
				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImageToCloudProfilesMapping, chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the infrastructure status cannot be decoded", func() {
				expectGetSecretCallToWork(c, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword)

				w.Spec.InfrastructureProviderStatus = &runtime.RawExtension{}

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImageToCloudProfilesMapping, chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the security group cannot be found", func() {
				expectGetSecretCallToWork(c, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword)

				w.Spec.InfrastructureProviderStatus = &runtime.RawExtension{
					Raw: encode(&apisopenstack.InfrastructureStatus{}),
				}

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImageToCloudProfilesMapping, chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the machine image for this cloud profile cannot be found", func() {
				expectGetSecretCallToWork(c, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword)

				cluster.CloudProfile.Name = "another-cloud-profile"

				workerDelegate = NewWorkerDelegate(c, scheme, decoder, machineImageToCloudProfilesMapping, chartApplier, "", w, cluster)

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

func expectGetSecretCallToWork(c *mockclient.MockClient, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword string) {
	c.EXPECT().
		Get(context.TODO(), gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
		DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
			secret.Data = map[string][]byte{
				openstack.DomainName: []byte(openstackDomainName),
				openstack.TenantName: []byte(openstackTenantName),
				openstack.UserName:   []byte(openstackUserName),
				openstack.Password:   []byte(openstackPassword),
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

func addNameAndSecretToMachineClass(class map[string]interface{}, openstackAuthURL, openstackDomainName, openstackTenantName, openstackUserName, openstackPassword, name string) {
	class["name"] = name
	class["secret"].(map[string]interface{})[openstack.AuthURL] = openstackAuthURL
	class["secret"].(map[string]interface{})[openstack.DomainName] = openstackDomainName
	class["secret"].(map[string]interface{})[openstack.TenantName] = openstackTenantName
	class["secret"].(map[string]interface{})[openstack.UserName] = openstackUserName
	class["secret"].(map[string]interface{})[openstack.Password] = openstackPassword
}
