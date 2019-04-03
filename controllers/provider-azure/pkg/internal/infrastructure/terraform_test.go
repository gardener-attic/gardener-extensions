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

package infrastructure

import (
	"testing"

	azurev1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/azure/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/internal"
	"github.com/gardener/gardener-extensions/pkg/controller"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func makeCluster(pods, services gardencorev1alpha1.CIDR, region string) *controller.Cluster {
	var (
		shoot = gardenv1beta1.Shoot{
			Spec: gardenv1beta1.ShootSpec{
				Cloud: gardenv1beta1.Cloud{
					Azure: &gardenv1beta1.AzureCloud{
						Networks: gardenv1beta1.AzureNetworks{
							K8SNetworks: gardencorev1alpha1.K8SNetworks{
								Pods:     &pods,
								Services: &services,
							},
						},
					},
				},
			},
		}
		cloudProfile = gardenv1beta1.CloudProfile{
			Spec: gardenv1beta1.CloudProfileSpec{
				Azure: &gardenv1beta1.AzureProfile{
					CountFaultDomains: []gardenv1beta1.AzureDomainCount{
						{Region: region, Count: 1},
					},
					CountUpdateDomains: []gardenv1beta1.AzureDomainCount{
						{Region: region, Count: 1},
					},
				},
			},
		}
	)

	return &controller.Cluster{
		Shoot:        &shoot,
		CloudProfile: &cloudProfile,
	}
}

func TestInfrastructure(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infrastructure Suite")
}

var _ = Describe("Terraform", func() {
	var (
		infra      *extensionsv1alpha1.Infrastructure
		config     *azurev1alpha1.InfrastructureConfig
		cluster    *controller.Cluster
		clientAuth *internal.ClientAuth
	)

	BeforeEach(func() {
		var (
			VNetName = "vnet"
			TestCIDR = gardencorev1alpha1.CIDR("10.1.0.0/16")
			VNetCIDR = TestCIDR
		)
		config = &azurev1alpha1.InfrastructureConfig{
			Networks: azurev1alpha1.NetworkConfig{
				VNet: &azurev1alpha1.VNet{
					Name: &VNetName,
					CIDR: &VNetCIDR,
				},
				Workers: TestCIDR,
			},
		}

		infra = &extensionsv1alpha1.Infrastructure{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo",
				Name:      "bar",
			},

			Spec: extensionsv1alpha1.InfrastructureSpec{
				Region: "eu-west-1",
				SecretRef: corev1.SecretReference{
					Namespace: "foo",
					Name:      "azure-credentials",
				},
				ProviderConfig: &runtime.RawExtension{
					Object: config,
				},
			},
		}

		cluster = makeCluster(gardencorev1alpha1.CIDR("11.0.0.0/16"), gardencorev1alpha1.CIDR("12.0.0.0/16"), infra.Spec.Region)
		clientAuth = &internal.ClientAuth{
			TenantID:       "tenant_id",
			ClientSecret:   "client_secret",
			ClientID:       "client_id",
			SubscriptionID: "subscription_id",
		}
	})

	Describe("#ComputeTerraformerChartValues", func() {
		It("should correctly compute the terraformer chart values", func() {
			values, err := ComputeTerraformerChartValues(infra, clientAuth, config, cluster)
			Expect(err).To(Not(HaveOccurred()))

			expectedValues := map[string]interface{}{
				"azure": map[string]interface{}{
					"subscriptionID":     clientAuth.SubscriptionID,
					"tenantID":           clientAuth.TenantID,
					"region":             infra.Spec.Region,
					"countUpdateDomains": cluster.CloudProfile.Spec.Azure.CountUpdateDomains[0].Count,
					"countFaultDomains":  cluster.CloudProfile.Spec.Azure.CountFaultDomains[0].Count,
				},
				"create": map[string]interface{}{
					"resourceGroup": true,
					"vnet":          false,
				},
				"resourceGroup": map[string]interface{}{
					"name": infra.Namespace,
					"vnet": map[string]interface{}{
						"name": *config.Networks.VNet.Name,
						"cidr": config.Networks.Workers,
					},
				},
				"clusterName": infra.Namespace,
				"networks": map[string]interface{}{
					"worker": config.Networks.Workers,
				},
				"outputKeys": map[string]interface{}{
					"resourceGroupName":   TerraformerOutputKeyResourceGroupName,
					"vnetName":            TerraformerOutputKeyVNetName,
					"subnetName":          TerraformerOutputKeySubnetName,
					"availabilitySetID":   TerraformerOutputKeyAvailabilitySetID,
					"availabilitySetName": TerraformerOutputKeyAvailabilitySetName,
					"routeTableName":      TerraformerOutputKeyRouteTableName,
					"securityGroupName":   TerraformerOutputKeySecurityGroupName,
				},
			}
			Expect(values).To(BeEquivalentTo(expectedValues))
		})
	})

	Describe("#StatusFromTerraformState", func() {
		var (
			vnetName, subnetName, routeTableName, availabilitySetID, availabilitySetName, securityGroupName, resourceGroupName string
			state                                                                                                              *TerraformState
		)

		BeforeEach(func() {
			vnetName = "vnet_name"
			subnetName = "subnet_name"
			routeTableName = "routTable_name"
			availabilitySetID, availabilitySetName = "as_id", "as_name"
			securityGroupName = "sg_name"
			resourceGroupName = "rg_name"
			state = &TerraformState{
				VNetName:            vnetName,
				SubnetName:          subnetName,
				RouteTableName:      routeTableName,
				AvailabilitySetID:   availabilitySetID,
				AvailabilitySetName: availabilitySetName,
				SecurityGroupName:   securityGroupName,
				ResourceGroupName:   resourceGroupName,
			}
		})

		It("should correctly compute the status", func() {
			status := StatusFromTerraformState(state)
			Expect(status).To(Equal(&azurev1alpha1.InfrastructureStatus{
				TypeMeta: StatusTypeMeta,
				ResourceGroup: &azurev1alpha1.ResourceGroup{
					Name: resourceGroupName,
				},
				RouteTables: []azurev1alpha1.RouteTable{
					{Name: routeTableName, Purpose: azurev1alpha1.PurposeNodes},
				},
				AvailabilitySets: []azurev1alpha1.AvailabilitySet{
					{Name: availabilitySetName, ID: availabilitySetID, Purpose: azurev1alpha1.PurposeNodes},
				},
				SecurityGroups: []azurev1alpha1.SecurityGroup{
					{Name: securityGroupName, Purpose: azurev1alpha1.PurposeNodes},
				},
				Networks: &azurev1alpha1.NetworkStatus{
					VNet: azurev1alpha1.VNet{
						Name: &vnetName,
					},
					Subnets: []azurev1alpha1.Subnet{
						{
							Purpose: azurev1alpha1.PurposeNodes,
							Name:    subnetName,
						},
					},
				},
			}))
		})
	})

})
