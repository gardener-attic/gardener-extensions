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
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	. "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/controller/infrastructure"
	mockclient "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/mock/provider-alicloud/alicloud/client"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		ctrl *gomock.Controller
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#GetVPCInfo", func() {
		It("should get info about the specified VPC", func() {
			var (
				client       = mockclient.NewMockVPC(ctrl)
				vpcID        = "vpcID"
				vpcCIDR      = "vpcCIDR"
				natGatewayID = "natGatewayID"
				sNATTableID1 = "sNATTableID1"
				sNATTableID2 = "sNATTableID2"
				sNATTableIDs = fmt.Sprintf("%s,%s", sNATTableID1, sNATTableID2)
			)

			describeVPCsReq := vpc.CreateDescribeVpcsRequest()
			describeVPCsReq.VpcId = vpcID

			describeNATGatewaysReq := vpc.CreateDescribeNatGatewaysRequest()
			describeNATGatewaysReq.VpcId = vpcID

			gomock.InOrder(
				client.EXPECT().DescribeVpcs(describeVPCsReq).Return(&vpc.DescribeVpcsResponse{
					Vpcs: vpc.Vpcs{
						Vpc: []vpc.Vpc{
							{CidrBlock: vpcCIDR},
						},
					},
				}, nil),

				client.EXPECT().DescribeNatGateways(describeNATGatewaysReq).Return(&vpc.DescribeNatGatewaysResponse{
					NatGateways: vpc.NatGateways{
						NatGateway: []vpc.NatGateway{
							{
								NatGatewayId: natGatewayID,
								SnatTableIds: vpc.SnatTableIdsInDescribeNatGateways{
									SnatTableId: []string{sNATTableID1, sNATTableID2},
								},
							},
						},
					},
				}, nil),
			)

			info, err := GetVPCInfo(client, vpcID)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(Equal(&VPCInfo{
				CIDR:         vpcCIDR,
				NATGatewayID: natGatewayID,
				SNATTableIDs: sNATTableIDs,
			}))
		})
	})
})
