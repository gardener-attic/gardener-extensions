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
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/v1alpha1"
	. "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/controller/infrastructure"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TerraformChartOps", func() {
	var (
		ctrl *gomock.Controller
		ops  = DefaultTerraformOps()
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#ComputeCreateVPCInitializerValues", func() {
		It("should compute the values from the config", func() {
			var (
				cidr               = "192.168.0.0/16"
				internetChargeType = "foo"
				config             = v1alpha1.InfrastructureConfig{
					Networks: v1alpha1.Networks{
						VPC: v1alpha1.VPC{
							CIDR: &cidr,
						},
					},
				}
			)

			Expect(ops.ComputeCreateVPCInitializerValues(&config, internetChargeType)).To(Equal(&InitializerValues{
				CreateVPC:          true,
				VPCID:              TerraformDefaultVPCID,
				VPCCIDR:            string(cidr),
				NATGatewayID:       TerraformDefaultNATGatewayID,
				SNATTableIDs:       TerraformDefaultSNATTableIDs,
				InternetChargeType: internetChargeType,
			}))
		})
	})

	Describe("#ComputeUseVPCInitializerValues", func() {
		It("should compute the values from the infra and config", func() {
			var (
				id           = "id"
				cidr         = "192.168.0.0/16"
				natGatewayID = "natGatewayID"
				sNATTableIDs = "sNATTableIDs"
				info         = VPCInfo{
					CIDR:         cidr,
					NATGatewayID: natGatewayID,
					SNATTableIDs: sNATTableIDs,
				}
				config = v1alpha1.InfrastructureConfig{
					Networks: v1alpha1.Networks{
						VPC: v1alpha1.VPC{
							ID: &id,
						},
					},
				}
			)

			Expect(ops.ComputeUseVPCInitializerValues(&config, &info)).To(Equal(&InitializerValues{
				CreateVPC:    false,
				VPCID:        id,
				VPCCIDR:      cidr,
				NATGatewayID: natGatewayID,
				SNATTableIDs: sNATTableIDs,
			}))
		})
	})

	Describe("#ComputeTerraformerChartValues", func() {
		It("should compute the terraformer chart values", func() {
			var (
				namespace    = "cluster-foo"
				region       = "region"
				sshPublicKey = "sshPublicKey"

				infra = extensionsv1alpha1.Infrastructure{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
					},
					Spec: extensionsv1alpha1.InfrastructureSpec{
						Region:       region,
						SSHPublicKey: []byte(sshPublicKey),
					},
				}

				zone1Name   = "zone1"
				zone1Worker = "192.168.0.0/16"

				zone2Name   = "zone2"
				zone2Worker = "192.169.0.0/16"

				config = v1alpha1.InfrastructureConfig{
					Networks: v1alpha1.Networks{
						Zones: []v1alpha1.Zone{
							{
								Name:   zone1Name,
								Worker: zone1Worker,
							},
							{
								Name:   zone2Name,
								Worker: zone2Worker,
							},
						},
					},
				}

				vpcCIDR            = "192.170.0.0/16"
				vpcID              = "vpcID"
				natGatewayID       = "natGatewayID"
				sNATTableIDs       = "sNATTableIDs"
				internetChargeType = "internetChargeType"
				values             = InitializerValues{
					CreateVPC:          true,
					VPCCIDR:            vpcCIDR,
					VPCID:              vpcID,
					NATGatewayID:       natGatewayID,
					SNATTableIDs:       sNATTableIDs,
					InternetChargeType: internetChargeType,
				}
			)

			Expect(ops.ComputeChartValues(&infra, &config, &values)).To(Equal(map[string]interface{}{
				"alicloud": map[string]interface{}{
					"region": region,
				},
				"create": map[string]interface{}{
					"vpc": true,
				},
				"vpc": map[string]interface{}{
					"cidr":               vpcCIDR,
					"id":                 vpcID,
					"natGatewayID":       natGatewayID,
					"snatTableID":        sNATTableIDs,
					"internetChargeType": internetChargeType,
				},
				"clusterName":  namespace,
				"sshPublicKey": sshPublicKey,
				"zones": []map[string]interface{}{
					{
						"name": zone1Name,
						"cidr": map[string]interface{}{
							"worker": zone1Worker,
						},
					},
					{
						"name": zone2Name,
						"cidr": map[string]interface{}{
							"worker": zone2Worker,
						},
					},
				},
				"outputKeys": map[string]interface{}{
					"vpcID":              TerraformerOutputKeyVPCID,
					"vpcCIDR":            TerraformerOutputKeyVPCCIDR,
					"securityGroupID":    TerraformerOutputKeySecurityGroupID,
					"keyPairName":        TerraformerOutputKeyKeyPairName,
					"vswitchNodesPrefix": TerraformerOutputKeyVSwitchNodesPrefix,
				},
			}))
		})
	})
})
