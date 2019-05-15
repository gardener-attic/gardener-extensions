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

package packet

import "path/filepath"

const (
	// Name is the name of the Packet provider controller.
	Name = "provider-packet"

	// TerraformerImageName is the name of the Terraformer image.
	TerraformerImageName = "terraformer"
	// CloudControllerManagerImageName is the name of the cloud-controller-manager image.
	CloudControllerManagerImageName = "cloud-controller-manager"
	// MachineControllerManagerImageName is the name of the MachineControllerManager image.
	MachineControllerManagerImageName = "machine-controller-manager"

	// AccessKeyID is a constant for the key in a cloud provider secret and backup secret that holds the Packet API token.
	APIToken = "apiToken"
	// ProjectID is a constant for the key in a cloud provider secret and backup secret that holds the Packet project id.
	ProjectID = "projectID"
	// PacketAPIKey is a constant for the key in a cloud provider secret and backup secret that holds the Packet API key.
	PacketAPIKey = "packetAPIKey"

	// TerraformerPurposeInfra is a constant for the complete Terraform setup with purpose 'infrastructure'.
	TerraformerPurposeInfra = "infra"
	// SSHKeyID key for accessing SSH key ID from outputs in terraform
	SSHKeyID = "key_pair_id"

	// MachineControllerManagerName is a constant for the name of the machine-controller-manager.
	MachineControllerManagerName = "machine-controller-manager"
)

var (
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("controllers", Name, "charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")
)

// Credentials stores Packet credentials.
type Credentials struct {
	APIToken  []byte
	ProjectID []byte
}
