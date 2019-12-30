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
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

type terraformOps struct{}

// DefaultTerraformOps returns the default TerraformChartOps.
func DefaultTerraformOps() TerraformChartOps {
	return terraformOps{}
}

// ComputeCreateVPCInitializerValues computes the InitializerValues to create a new VPC.
func (terraformOps) ComputeCreateVPCInitializerValues(config *v1alpha1.InfrastructureConfig, internetChargeType string) *InitializerValues {
	return &InitializerValues{
		CreateVPC:          true,
		VPCID:              TerraformDefaultVPCID,
		VPCCIDR:            string(*config.Networks.VPC.CIDR),
		NATGatewayID:       TerraformDefaultNATGatewayID,
		SNATTableIDs:       TerraformDefaultSNATTableIDs,
		InternetChargeType: internetChargeType,
	}
}

// ComputeUseVPCInitializerValues computes the InitializerValues to use an existing VPC.
func (terraformOps) ComputeUseVPCInitializerValues(config *v1alpha1.InfrastructureConfig, info *VPCInfo) *InitializerValues {
	return &InitializerValues{
		CreateVPC:          false,
		VPCID:              *config.Networks.VPC.ID,
		VPCCIDR:            info.CIDR,
		NATGatewayID:       info.NATGatewayID,
		SNATTableIDs:       info.SNATTableIDs,
		InternetChargeType: info.InternetChargeType,
	}
}

// ComputeTerraformerChartValues computes the values necessary for the infrastructure Terraform chart.
func (terraformOps) ComputeChartValues(
	infra *extensionsv1alpha1.Infrastructure,
	config *v1alpha1.InfrastructureConfig,
	values *InitializerValues,
) map[string]interface{} {
	zones := make([]map[string]interface{}, 0, len(config.Networks.Zones))
	for _, zone := range config.Networks.Zones {
		workersCIDR := zone.Workers
		// Backwards compatibility - remove this code in a future version.
		if workersCIDR == "" {
			workersCIDR = zone.Worker
		}
		zones = append(zones, map[string]interface{}{
			"name": zone.Name,
			"cidr": map[string]interface{}{
				"workers": string(workersCIDR),
			},
		})
	}

	return map[string]interface{}{
		"alicloud": map[string]interface{}{
			"region": infra.Spec.Region,
		},
		"create": map[string]interface{}{
			"vpc": values.CreateVPC,
		},
		"vpc": map[string]interface{}{
			"cidr":               values.VPCCIDR,
			"id":                 values.VPCID,
			"natGatewayID":       values.NATGatewayID,
			"snatTableID":        values.SNATTableIDs,
			"internetChargeType": values.InternetChargeType,
		},
		"clusterName":  infra.Namespace,
		"sshPublicKey": string(infra.Spec.SSHPublicKey),
		"zones":        zones,
		"outputKeys": map[string]interface{}{
			"vpcID":              TerraformerOutputKeyVPCID,
			"vpcCIDR":            TerraformerOutputKeyVPCCIDR,
			"securityGroupID":    TerraformerOutputKeySecurityGroupID,
			"keyPairName":        TerraformerOutputKeyKeyPairName,
			"vswitchNodesPrefix": TerraformerOutputKeyVSwitchNodesPrefix,
		},
	}
}
