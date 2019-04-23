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

package infrastructure_test

import (
	"context"
	gcpv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal"
	. "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/infrastructure"
	mockgcpclient "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/mock/client"
	mockinfra "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/mock/infrastructure"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/test"
	infratest "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/test/infrastructure"
	"github.com/gardener/gardener-extensions/pkg/controller"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockterraformer "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/gardener/terraformer"
	mockchartrenderer "github.com/gardener/gardener-extensions/pkg/mock/gardener/chartrenderer"
	testutil "github.com/gardener/gardener-extensions/pkg/util/test"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	gardenterraformer "github.com/gardener/gardener/pkg/operation/terraformer"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

var _ = Describe("Actuator Suite", func() {
	var (
		ctrl *gomock.Controller

		restConfig    *rest.Config
		c             *mockclient.MockClient
		chartRenderer *mockchartrenderer.MockInterface

		podNetwork     gardencorev1alpha1.CIDR
		serviceNetwork gardencorev1alpha1.CIDR
		cluster        controller.Cluster

		config                      gcpv1alpha1.InfrastructureConfig
		secretNamespace, secretName string
		projectID                   string
		serviceAccount              internal.ServiceAccount
		serviceAccountData          []byte
		secret                      corev1.Secret
		secretRef                   corev1.SecretReference
		infra                       extensionsv1alpha1.Infrastructure

		vpcName             string
		subnetNodes         string
		serviceAccountEmail string
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		restConfig = &rest.Config{}
		c = mockclient.NewMockClient(ctrl)
		chartRenderer = mockchartrenderer.NewMockInterface(ctrl)

		podNetwork = gardencorev1alpha1.CIDR("192.169.0.0/16")
		serviceNetwork = gardencorev1alpha1.CIDR("192.170.0.0/16")
		cluster = controller.Cluster{
			Shoot: &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						GCP: &gardenv1beta1.GCPCloud{
							Networks: gardenv1beta1.GCPNetworks{
								K8SNetworks: gardencorev1alpha1.K8SNetworks{
									Pods:     &podNetwork,
									Services: &serviceNetwork,
								},
							},
						},
					},
				},
			},
		}
		config = gcpv1alpha1.InfrastructureConfig{}
		secretNamespace, secretName = "secretNamespace", "secretName"
		secretRef = corev1.SecretReference{Namespace: secretNamespace, Name: secretName}
		projectID = "projectID"
		serviceAccountData = test.MkServiceAccountData(projectID)
		serviceAccount = internal.ServiceAccount{Raw: serviceAccountData, ProjectID: projectID}
		secret = *test.MkServiceAccountSecret(secretNamespace, secretName, serviceAccountData)
		infra = extensionsv1alpha1.Infrastructure{
			Spec: extensionsv1alpha1.InfrastructureSpec{
				ProviderConfig: &runtime.RawExtension{
					Object: &config,
				},
				SecretRef: secretRef,
			},
		}

		vpcName = "vpc"
		subnetNodes = "subnetNodes"
		serviceAccountEmail = "serviceAccountEmail"
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Actuator", func() {
		Describe("#Delete", func() {
			It("should correctly delete the infrastructure", func() {
				var (
					ctx = context.TODO()

					actuator = &Actuator{
						RESTConfig:    restConfig,
						Client:        c,
						ChartRenderer: chartRenderer,
					}

					gcpClient   = mockgcpclient.NewMockInterface(ctrl)
					terraformer = mockterraformer.NewMockTerraformer(ctrl)

					newGCPClientFromServiceAccount = mockinfra.NewMockNewGCPClientFromServiceAccount(ctrl)
					newTerraformer                 = mockinfra.NewMockNewTerraformer(ctrl)
				)

				gomock.InOrder(
					test.ExpectGetServiceAccountSecret(ctx, c, &secret),
					newGCPClientFromServiceAccount.EXPECT().Do(ctx, serviceAccountData).Return(gcpClient, nil),
					newTerraformer.EXPECT().Do(restConfig, &serviceAccount, TerraformerPurpose, infra.Namespace, infra.Name).Return(terraformer, nil),

					terraformer.EXPECT().ConfigExists().Return(false, nil),
					terraformer.EXPECT().Destroy(),
				)

				defer testutil.WithVars(
					&NewGCPClientFromServiceAccount, newGCPClientFromServiceAccount.Do,
					&NewTerraformer, newTerraformer.Do,
				)()

				Expect(actuator.Delete(ctx, &infra, nil)).To(Succeed())
			})
		})

		It("should also cleanup kubernetes artifacts if a configuration exists", func() {
			var (
				ctx = context.TODO()

				actuator = &Actuator{
					RESTConfig:    restConfig,
					Client:        c,
					ChartRenderer: chartRenderer,
				}

				gcpClient   = mockgcpclient.NewMockInterface(ctrl)
				terraformer = mockterraformer.NewMockTerraformer(ctrl)

				newGCPClientFromServiceAccount  = mockinfra.NewMockNewGCPClientFromServiceAccount(ctrl)
				newTerraformer                  = mockinfra.NewMockNewTerraformer(ctrl)
				cleanupKubernetesCloudArtifacts = mockinfra.NewMockCleanupKubernetesCloudArtifacts(ctrl)
			)

			gomock.InOrder(
				test.ExpectGetServiceAccountSecret(ctx, c, &secret),
				newGCPClientFromServiceAccount.EXPECT().Do(ctx, serviceAccountData).Return(gcpClient, nil),
				newTerraformer.EXPECT().Do(restConfig, &serviceAccount, TerraformerPurpose, infra.Namespace, infra.Name).Return(terraformer, nil),

				terraformer.EXPECT().ConfigExists().Return(true, nil),
				terraformer.EXPECT().GetStateOutputVariables(
					TerraformerOutputKeyVPCName,
					TerraformerOutputKeySubnetNodes,
					TerraformerOutputKeyServiceAccountEmail,
				).Return(infratest.MkTerraformerOutputVariables(vpcName, subnetNodes, serviceAccountEmail, nil), nil),
				cleanupKubernetesCloudArtifacts.EXPECT().Do(ctx, gcpClient, projectID, vpcName),

				terraformer.EXPECT().Destroy(),
			)

			defer testutil.WithVars(
				&NewGCPClientFromServiceAccount, newGCPClientFromServiceAccount.Do,
				&NewTerraformer, newTerraformer.Do,
				&CleanupKubernetesCloudArtifacts, cleanupKubernetesCloudArtifacts.Do,
			)()

			Expect(actuator.Delete(ctx, &infra, nil)).To(Succeed())
		})

		Describe("#Reconcile", func() {
			It("should correctly reconcile an infrastructure", func() {
				var (
					ctx = context.TODO()

					actuator = &Actuator{
						RESTConfig:    restConfig,
						Client:        c,
						ChartRenderer: chartRenderer,
					}

					terraformer = mockterraformer.NewMockTerraformer(ctrl)

					newTerraformer                = mockinfra.NewMockNewTerraformer(ctrl)
					terraformerDefaultInitializer = mockinfra.NewMockTerraformerDefaultInitializer(ctrl)
					terraformerInitializer        = mockinfra.NewMockTerraformerInitializer(ctrl)
				)

				gomock.InOrder(
					test.ExpectGetServiceAccountSecret(ctx, c, &secret),

					chartRenderer.EXPECT().Render(InfraChartPath, InfraChartName, infra.Namespace, ComputeTerraformerChartValues(&infra, &serviceAccount, &config, &cluster)).
						Return(&chartrenderer.RenderedChart{}, nil),

					newTerraformer.EXPECT().Do(restConfig, &serviceAccount, TerraformerPurpose, infra.Namespace, infra.Name).Return(terraformer, nil),

					terraformerDefaultInitializer.EXPECT().Do(c, "", "", []byte{}).Return(terraformerInitializer.Do),
					terraformer.EXPECT().InitializeWith(gomock.Any()).Do(func(init gardenterraformer.Initializer) error { return init(nil) }).Return(terraformer),
					terraformerInitializer.EXPECT().Do(nil),
					terraformer.EXPECT().Apply(),

					terraformer.EXPECT().GetStateOutputVariables(
						TerraformerOutputKeyVPCName,
						TerraformerOutputKeySubnetNodes,
						TerraformerOutputKeyServiceAccountEmail,
					).Return(infratest.MkTerraformerOutputVariables(vpcName, subnetNodes, serviceAccountEmail, nil), nil),

					c.EXPECT().Status().Return(c),
					c.EXPECT().Get(ctx, kutil.Key(infra.Namespace, infra.Name), &infra),
					c.EXPECT().Update(ctx, &infra).Do(func(_ context.Context, actual *extensionsv1alpha1.Infrastructure) {
						Expect(actual.Status.ProviderStatus.Object).To(Equal(StatusFromTerraformState(&TerraformState{
							VPCName:             vpcName,
							ServiceAccountEmail: serviceAccountEmail,
							SubnetNodes:         subnetNodes,
							SubnetInternal:      nil,
						})))
					}),
				)

				defer testutil.WithVars(
					&NewTerraformer, newTerraformer.Do,
					&TerraformerDefaultInitializer, terraformerDefaultInitializer.Do,
				)()

				Expect(actuator.Reconcile(ctx, &infra, &cluster)).To(Succeed())
			})
		})
	})
})
