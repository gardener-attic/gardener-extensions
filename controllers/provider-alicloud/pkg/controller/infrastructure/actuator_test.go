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
	"fmt"
	"time"

	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud"
	alicloudclient "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud/client"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/install"
	alicloudv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/config"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/controller/common"
	. "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/controller/infrastructure"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/imagevector"
	mockalicloudclient "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/mock/provider-alicloud/alicloud/client"
	mockinfrastructure "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/mock/provider-alicloud/controller/infrastructure"
	"github.com/gardener/gardener-extensions/pkg/controller"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockchartrenderer "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/gardener/chartrenderer"
	mockterraformer "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/terraformer"
	mockgardenerchartrenderer "github.com/gardener/gardener-extensions/pkg/mock/gardener/chartrenderer"
	"github.com/gardener/gardener-extensions/pkg/mock/go-logr/logr"
	"github.com/gardener/gardener-extensions/pkg/util/chart"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/manifest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

func ExpectInject(ok bool, err error) {
	Expect(err).NotTo(HaveOccurred())
	Expect(ok).To(BeTrue(), "no injection happened")
}

func ExpectEncode(data []byte, err error) []byte {
	Expect(err).NotTo(HaveOccurred())
	Expect(data).NotTo(BeNil())
	return data
}

func mkManifest(name string, content string) manifest.Manifest {
	return manifest.Manifest{
		Name:    fmt.Sprintf("/templates/%s", name),
		Content: content,
	}
}

var _ = Describe("Actuator", func() {
	var (
		ctrl       *gomock.Controller
		scheme     *runtime.Scheme
		serializer runtime.Serializer
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		scheme = runtime.NewScheme()
		install.Install(scheme)
		Expect(controller.AddToScheme(scheme)).To(Succeed())
		serializer = json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Actuator", func() {
		Describe("#Reconcile", func() {
			It("should correctly reconcile the infrastructure", func() {
				var (
					ctx                      = context.TODO()
					logger                   = logr.NewMockLogger(ctrl)
					newAlicloudClientFactory = mockalicloudclient.NewMockClientFactory(ctrl)
					alicloudClientFactory    = mockalicloudclient.NewMockFactory(ctrl)
					vpcClient                = mockalicloudclient.NewMockVPC(ctrl)
					terraformerFactory       = mockterraformer.NewMockFactory(ctrl)
					terraformer              = mockterraformer.NewMockTerraformer(ctrl)
					shootECSClient           = mockalicloudclient.NewMockECS(ctrl)
					shootSTSClient           = mockalicloudclient.NewMockSTS(ctrl)
					chartRendererFactory     = mockchartrenderer.NewMockFactory(ctrl)
					terraformChartOps        = mockinfrastructure.NewMockTerraformChartOps(ctrl)
					machineImageMapping      = []config.MachineImage{}
					actuator                 = NewActuatorWithDeps(
						logger,
						newAlicloudClientFactory,
						alicloudClientFactory,
						terraformerFactory,
						chartRendererFactory,
						terraformChartOps,
						machineImageMapping,
						nil,
					)
					c           = mockclient.NewMockClient(ctrl)
					initializer = mockterraformer.NewMockInitializer(ctrl)
					restConfig  rest.Config

					chartRenderer = mockgardenerchartrenderer.NewMockInterface(ctrl)

					cidr   = "192.168.0.0/16"
					config = alicloudv1alpha1.InfrastructureConfig{
						Networks: alicloudv1alpha1.Networks{
							VPC: alicloudv1alpha1.VPC{
								CIDR: &cidr,
							},
						},
					}
					configYAML      = ExpectEncode(runtime.Encode(serializer, &config))
					secretNamespace = "secretns"
					secretName      = "secret"
					region          = "region"
					infra           = extensionsv1alpha1.Infrastructure{
						Spec: extensionsv1alpha1.InfrastructureSpec{
							ProviderConfig: &runtime.RawExtension{
								Raw: configYAML,
							},
							Region: region,
							SecretRef: corev1.SecretReference{
								Namespace: secretNamespace,
								Name:      secretName,
							},
						},
					}
					accessKeyID     = "accessKeyID"
					accessKeySecret = "accessKeySecret"
					cluster         = controller.Cluster{
						Shoot: &gardencorev1alpha1.Shoot{
							Spec: gardencorev1alpha1.ShootSpec{
								Region: region,
							},
						},
					}

					initializerValues = InitializerValues{}
					chartValues       = map[string]interface{}{}

					mainContent      = "main"
					variablesContent = "variables"
					tfVarsContent    = "tfVars"

					vpcID           = "vpcID"
					vpcCIDRString   = "vpcCIDR"
					natGatewayID    = "natGatewayID"
					securityGroupID = "sgID"
					keyPairName     = "keyPairName"
				)

				describeNATGatewaysReq := vpc.CreateDescribeNatGatewaysRequest()
				describeNATGatewaysReq.VpcId = vpcID

				gomock.InOrder(
					chartRendererFactory.EXPECT().NewForConfig(&restConfig).Return(chartRenderer, nil),

					c.EXPECT().Get(ctx, client.ObjectKey{Namespace: secretNamespace, Name: secretName}, gomock.AssignableToTypeOf(&corev1.Secret{})).
						SetArg(2, corev1.Secret{
							Data: map[string][]byte{
								alicloud.AccessKeyID:     []byte(accessKeyID),
								alicloud.AccessKeySecret: []byte(accessKeySecret),
							},
						}),

					terraformerFactory.EXPECT().NewForConfig(gomock.Any(), &restConfig, TerraformerPurpose, infra.Namespace, infra.Name, imagevector.TerraformerImage()).
						Return(terraformer, nil),

					terraformer.EXPECT().SetVariablesEnvironment(map[string]string{
						common.TerraformVarAccessKeyID:     accessKeyID,
						common.TerraformVarAccessKeySecret: accessKeySecret,
					}).Return(terraformer),
					terraformer.EXPECT().SetActiveDeadlineSeconds(int64(630)).Return(terraformer),
					terraformer.EXPECT().SetDeadlineCleaning(5*time.Minute).Return(terraformer),
					terraformer.EXPECT().SetDeadlinePod(15*time.Minute).Return(terraformer),

					alicloudClientFactory.EXPECT().NewVPC(region, accessKeyID, accessKeySecret).Return(vpcClient, nil),

					terraformer.EXPECT().GetStateOutputVariables(TerraformerOutputKeyVPCID).
						Return(map[string]string{
							TerraformerOutputKeyVPCID: vpcID,
						}, nil),

					vpcClient.EXPECT().DescribeNatGateways(describeNATGatewaysReq).Return(&vpc.DescribeNatGatewaysResponse{
						NatGateways: vpc.NatGateways{
							NatGateway: []vpc.NatGateway{
								{
									NatGatewayId: natGatewayID,
								},
							},
						},
					}, nil),

					terraformChartOps.EXPECT().ComputeCreateVPCInitializerValues(&config, alicloudclient.DefaultInternetChargeType).Return(&initializerValues),
					terraformChartOps.EXPECT().ComputeChartValues(&infra, &config, &initializerValues).Return(chartValues),

					chartRenderer.EXPECT().Render(
						alicloud.InfraChartPath,
						alicloud.InfraRelease,
						infra.Namespace,
						chartValues,
					).Return(&chartrenderer.RenderedChart{
						Manifests: []manifest.Manifest{
							mkManifest(chart.TerraformMainTFFilename, mainContent),
							mkManifest(chart.TerraformVariablesTFFilename, variablesContent),
							mkManifest(chart.TerraformTFVarsFilename, tfVarsContent),
						},
					}, nil),

					terraformerFactory.EXPECT().DefaultInitializer(c, mainContent, variablesContent, []byte(tfVarsContent)).Return(initializer),

					terraformer.EXPECT().InitializeWith(initializer).Return(terraformer),

					terraformer.EXPECT().Apply(),

					c.EXPECT().Get(ctx, client.ObjectKey{Namespace: secretNamespace, Name: secretName}, gomock.AssignableToTypeOf(&corev1.Secret{})).
						SetArg(2, corev1.Secret{
							Data: map[string][]byte{
								alicloud.AccessKeyID:     []byte(accessKeyID),
								alicloud.AccessKeySecret: []byte(accessKeySecret),
							},
						}),
					logger.EXPECT().Info("Creating Alicloud ECS client for Shoot", "infrastructure", infra.Name),
					newAlicloudClientFactory.EXPECT().NewECSClient(ctx, region, accessKeyID, accessKeySecret).Return(shootECSClient, nil),
					logger.EXPECT().Info("Creating Alicloud STS client for Shoot", "infrastructure", infra.Name),
					newAlicloudClientFactory.EXPECT().NewSTSClient(ctx, region, accessKeyID, accessKeySecret).Return(shootSTSClient, nil),
					shootSTSClient.EXPECT().GetAccountIDFromCallerIdentity(ctx).Return("", nil),
					logger.EXPECT().Info("Sharing customized image with Shoot's Alicloud account from Seed", "infrastructure", infra.Name),

					terraformer.EXPECT().GetStateOutputVariables(TerraformerOutputKeyVPCID, TerraformerOutputKeyVPCCIDR, TerraformerOutputKeySecurityGroupID, TerraformerOutputKeyKeyPairName).
						Return(map[string]string{
							TerraformerOutputKeyVPCID:           vpcID,
							TerraformerOutputKeyVPCCIDR:         vpcCIDRString,
							TerraformerOutputKeySecurityGroupID: securityGroupID,
							TerraformerOutputKeyKeyPairName:     keyPairName,
						}, nil),

					c.EXPECT().Status().Return(c),
					c.EXPECT().Get(ctx, client.ObjectKey{Namespace: infra.Namespace, Name: infra.Name}, &infra),

					c.EXPECT().Update(ctx, &infra),
				)

				ExpectInject(inject.ClientInto(c, actuator))
				ExpectInject(inject.SchemeInto(scheme, actuator))
				ExpectInject(inject.ConfigInto(&restConfig, actuator))

				Expect(actuator.Reconcile(ctx, &infra, &cluster)).To(Succeed())
				Expect(infra.Status.ProviderStatus.Object).To(Equal(&alicloudv1alpha1.InfrastructureStatus{
					TypeMeta: StatusTypeMeta,
					VPC: alicloudv1alpha1.VPCStatus{
						ID: vpcID,
						SecurityGroups: []alicloudv1alpha1.SecurityGroup{
							{
								Purpose: alicloudv1alpha1.PurposeNodes,
								ID:      securityGroupID,
							},
						},
					},
					KeyPairName: keyPairName,
				}))
			})
		})
	})
})
