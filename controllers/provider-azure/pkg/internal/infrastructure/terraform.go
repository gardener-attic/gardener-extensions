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
	"fmt"
	"path/filepath"

	azurev1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/azure/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/internal"
	"github.com/gardener/gardener-extensions/pkg/controller"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/operation/terraformer"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TerraformerPurpose is the terraformer infrastructure purpose.
	TerraformerPurpose = "infra"

	// TerraformerOutputKeyResourceGroupName is the key for the resourceGroupName output
	TerraformerOutputKeyResourceGroupName = "resourceGroupName"
	// TerraformerOutputKeyVNetName is the key for the vnetName output
	TerraformerOutputKeyVNetName = "vnetName"
	// TerraformerOutputKeySubnetName is the key for the subnetName output
	TerraformerOutputKeySubnetName = "subnetName"
	// TerraformerOutputKeyAvailabilitySetID is the key for the availabilitySetID output
	TerraformerOutputKeyAvailabilitySetID = "availabilitySetID"
	// TerraformerOutputKeyAvailabilitySetName is the key for the availabilitySetName output
	TerraformerOutputKeyAvailabilitySetName = "availabilitySetName"
	// TerraformerOutputKeyRouteTableName is the key for the routeTableName output
	TerraformerOutputKeyRouteTableName = "routeTableName"
	// TerraformerOutputKeySecurityGroupName is the key for the securityGroupName output
	TerraformerOutputKeySecurityGroupName = "securityGroupName"
)

var (
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("controllers", "provider-azure", "charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")

	// StatusTypeMeta is the TypeMeta of the Azure InfrastructureStatus
	StatusTypeMeta = metav1.TypeMeta{
		APIVersion: azurev1alpha1.SchemeGroupVersion.String(),
		Kind:       "InfrastructureStatus",
	}
)

// ComputeTerraformerChartValues computes the values for the Azure Terraformer chart.
func ComputeTerraformerChartValues(infra *extensionsv1alpha1.Infrastructure, clientAuth *internal.ClientAuth,
	config *azurev1alpha1.InfrastructureConfig, cluster *controller.Cluster) (map[string]interface{}, error) {
	var (
		createResourceGroup   = true
		createVNet            = true
		createAvailabilitySet = false
		resourceGroupName     = infra.Namespace
		vnetName              = infra.Namespace
		vnetCIDR              = config.Networks.Workers

		findDomainCountByRegion = func(region string, domainCounts []v1beta1.AzureDomainCount) (v1beta1.AzureDomainCount, error) {
			for _, domainCount := range domainCounts {
				if domainCount.Region == region {
					return domainCount, nil
				}
			}
			return v1beta1.AzureDomainCount{}, fmt.Errorf("could not find a domain count for region %s", region)
		}

		azure = map[string]interface{}{
			"subscriptionID": clientAuth.SubscriptionID,
			"tenantID":       clientAuth.TenantID,
			"region":         infra.Spec.Region,
		}
		outputKeys = map[string]interface{}{
			"resourceGroupName": TerraformerOutputKeyResourceGroupName,
			"vnetName":          TerraformerOutputKeyVNetName,
			"subnetName":        TerraformerOutputKeySubnetName,
			"routeTableName":    TerraformerOutputKeyRouteTableName,
			"securityGroupName": TerraformerOutputKeySecurityGroupName,
		}
	)
	// check if we should use an existing ResourceGroup or create a new one
	if config.ResourceGroup != nil {
		createResourceGroup = false
		resourceGroupName = config.ResourceGroup.Name
	}

	// check if we should use an existing ResourceGroup or create a new one
	if config.Networks.VNet.Name != nil {
		createVNet = false
		vnetName = *config.Networks.VNet.Name
	}
	if config.Networks.VNet.CIDR != nil {
		vnetCIDR = *config.Networks.VNet.CIDR
	}

	// If the cluster is zoned, then we don't need to create an AvailabilitySet.
	if !config.Zoned && cluster.CloudProfile.Spec.Azure != nil {
		createAvailabilitySet = true
		outputKeys["availabilitySetID"] = TerraformerOutputKeyAvailabilitySetID
		outputKeys["availabilitySetName"] = TerraformerOutputKeyAvailabilitySetName

		countUpdateDomains, err := findDomainCountByRegion(infra.Spec.Region, cluster.CloudProfile.Spec.Azure.CountUpdateDomains)
		if err != nil {
			return nil, err
		}
		azure["countUpdateDomains"] = countUpdateDomains.Count

		countFaultDomains, err := findDomainCountByRegion(infra.Spec.Region, cluster.CloudProfile.Spec.Azure.CountFaultDomains)
		if err != nil {
			return nil, err
		}
		azure["countFaultDomains"] = countFaultDomains.Count
	}

	return map[string]interface{}{
		"azure": azure,
		"create": map[string]interface{}{
			"resourceGroup":   createResourceGroup,
			"vnet":            createVNet,
			"availabilitySet": createAvailabilitySet,
		},
		"resourceGroup": map[string]interface{}{
			"name": resourceGroupName,
			"vnet": map[string]interface{}{
				"name": vnetName,
				"cidr": vnetCIDR,
			},
		},
		"clusterName": infra.Namespace,
		"networks": map[string]interface{}{
			"worker": config.Networks.Workers,
		},
		"outputKeys": outputKeys,
	}, nil
}

// RenderTerraformerChart renders the azure-infra chart with the given values.
func RenderTerraformerChart(renderer chartrenderer.Interface, infra *extensionsv1alpha1.Infrastructure, clientAuth *internal.ClientAuth,
	config *azurev1alpha1.InfrastructureConfig, cluster *controller.Cluster) (*TerraformFiles, error) {
	values, err := ComputeTerraformerChartValues(infra, clientAuth, config, cluster)
	if err != nil {
		return nil, err
	}

	release, err := renderer.Render(filepath.Join(InternalChartsPath, "azure-infra"), "azure-infra", infra.Namespace, values)
	if err != nil {
		return nil, err
	}

	return &TerraformFiles{
		Main:      release.FileContent("main.tf"),
		Variables: release.FileContent("variables.tf"),
		TFVars:    []byte(release.FileContent("terraform.tfvars")),
	}, nil
}

// TerraformFiles are the files that have been rendered from the infrastructure chart.
type TerraformFiles struct {
	Main      string
	Variables string
	TFVars    []byte
}

// TerraformState is the Terraform state for an infrastructure.
type TerraformState struct {
	// VPCName is the name of the VNet created for an infrastructure.
	VNetName string
	// ResourceGroupName is the name of the resource group.
	ResourceGroupName string
	// AvailabilitySetID is the ID for the created availability set.
	AvailabilitySetID string
	// AvailabilitySetName the ID for the created availability set .
	AvailabilitySetName string
	// SubnetName is the name of the created subnet.
	SubnetName string
	// RouteTableName is the name of the route table.
	RouteTableName string
	// SecurityGroupName is the name of the security group.
	SecurityGroupName string
}

// ExtractTerraformState extracts the TerraformState from the given Terraformer.
func ExtractTerraformState(tf *terraformer.Terraformer, config *azurev1alpha1.InfrastructureConfig) (*TerraformState, error) {
	var outputKeys = []string{
		TerraformerOutputKeyResourceGroupName,
		TerraformerOutputKeyRouteTableName,
		TerraformerOutputKeySecurityGroupName,
		TerraformerOutputKeySubnetName,
		TerraformerOutputKeyVNetName,
	}

	if !config.Zoned {
		outputKeys = append(outputKeys, TerraformerOutputKeyAvailabilitySetID, TerraformerOutputKeyAvailabilitySetName)
	}

	vars, err := tf.GetStateOutputVariables(outputKeys...)
	if err != nil {
		return nil, err
	}

	var tfState = TerraformState{
		VNetName:          vars[TerraformerOutputKeyVNetName],
		ResourceGroupName: vars[TerraformerOutputKeyResourceGroupName],
		RouteTableName:    vars[TerraformerOutputKeyRouteTableName],
		SecurityGroupName: vars[TerraformerOutputKeySecurityGroupName],
		SubnetName:        vars[TerraformerOutputKeySubnetName],
	}

	if !config.Zoned {
		tfState.AvailabilitySetID = vars[TerraformerOutputKeyAvailabilitySetID]
		tfState.AvailabilitySetName = vars[TerraformerOutputKeyAvailabilitySetName]
	}
	return &tfState, nil
}

// StatusFromTerraformState computes an InfrastructureStatus from the given
// Terraform variables.
func StatusFromTerraformState(state *TerraformState) *azurev1alpha1.InfrastructureStatus {
	var tfState = azurev1alpha1.InfrastructureStatus{
		TypeMeta: StatusTypeMeta,
		ResourceGroup: azurev1alpha1.ResourceGroup{
			Name: state.ResourceGroupName,
		},
		Networks: azurev1alpha1.NetworkStatus{
			VNet: azurev1alpha1.VNetStatus{
				Name: state.VNetName,
			},
			Subnets: []azurev1alpha1.Subnet{
				{
					Purpose: azurev1alpha1.PurposeNodes,
					Name:    state.SubnetName,
				},
			},
		},
		AvailabilitySets: []azurev1alpha1.AvailabilitySet{},
		RouteTables: []azurev1alpha1.RouteTable{
			{Purpose: azurev1alpha1.PurposeNodes, Name: state.RouteTableName},
		},
		SecurityGroups: []azurev1alpha1.SecurityGroup{
			{Name: state.SecurityGroupName, Purpose: azurev1alpha1.PurposeNodes},
		},
	}

	// If no AvailabilitySet was created then the Shoot uses zones.
	if state.AvailabilitySetID == "" && state.AvailabilitySetName == "" {
		tfState.Zoned = true
	} else {
		tfState.AvailabilitySets = append(tfState.AvailabilitySets, azurev1alpha1.AvailabilitySet{
			Name:    state.AvailabilitySetName,
			ID:      state.AvailabilitySetID,
			Purpose: azurev1alpha1.PurposeNodes,
		})
	}

	return &tfState
}

// ComputeStatus computes the status based on the Terraformer and the given InfrastructureConfig.
func ComputeStatus(tf *terraformer.Terraformer, config *azurev1alpha1.InfrastructureConfig) (*azurev1alpha1.InfrastructureStatus, error) {
	state, err := ExtractTerraformState(tf, config)
	if err != nil {
		return nil, err
	}

	return StatusFromTerraformState(state), nil
}
