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
	gcpv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/infrastructure"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var (
	serializer = json.NewYAMLSerializer(json.DefaultMetaFactory, internal.Scheme, internal.Scheme)
)

// MkRawInfrastructureConfig encodes the given config into its raw representation.
func MkRawInfrastructureConfig(config *gcpv1alpha1.InfrastructureConfig) ([]byte, error) {
	return runtime.Encode(serializer, config)
}

// MkTerraformerOutputVariables creates a map with the terraform output variables for infrastructure.
func MkTerraformerOutputVariables(vpcName, subnetNodes, serviceAccountEmail string, subnetInternal *string) map[string]string {
	out := map[string]string{
		infrastructure.TerraformerOutputKeyVPCName:             vpcName,
		infrastructure.TerraformerOutputKeySubnetNodes:         subnetNodes,
		infrastructure.TerraformerOutputKeyServiceAccountEmail: serviceAccountEmail,
	}
	if subnetInternal != nil {
		out[infrastructure.TerraformerOutputKeySubnetInternal] = *subnetInternal
	}
	return out
}
