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

package client

import (
	alicloudvpc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
)

// DefaultInternetChargeType is used for EIP
const DefaultInternetChargeType = "PayByTraffic"

// VPC is the interface to the Alicloud VPC service.
type VPC interface {
	// DescribeVpcs describes the VPCs for the request.
	DescribeVpcs(req *alicloudvpc.DescribeVpcsRequest) (*alicloudvpc.DescribeVpcsResponse, error)
	// DescribeNatGateways describes the NAT gateways for the request.
	DescribeNatGateways(req *alicloudvpc.DescribeNatGatewaysRequest) (*alicloudvpc.DescribeNatGatewaysResponse, error)
	// DescribeEipAddresses describes the EIP addresses for the request.
	DescribeEipAddresses(req *alicloudvpc.DescribeEipAddressesRequest) (*alicloudvpc.DescribeEipAddressesResponse, error)
}

// Factory is the factory to instantiate Alicloud clients.
type Factory interface {
	// NewVPC creates a new VPC client from the given credentials and region.
	NewVPC(region, accessKeyID, accessKeySecret string) (VPC, error)
}
