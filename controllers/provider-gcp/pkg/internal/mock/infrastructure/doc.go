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

//go:generate mockgen -package=infrastructure -destination=funcs.go github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/mock/infrastructure NewGCPClientFromServiceAccount,NewTerraformer,CleanupKubernetesCloudArtifacts,TerraformerDefaultInitializer,TerraformerInitializer

package infrastructure

import (
	"context"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal"
	gcpclient "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/client"
	"github.com/gardener/gardener-extensions/pkg/gardener/terraformer"
	gardenterraformer "github.com/gardener/gardener/pkg/operation/terraformer"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewGCPClientFromServiceAccount is a mock for the NewGCPClientFromServiceAccount function.
type NewGCPClientFromServiceAccount interface {
	Do(ctx context.Context, serviceAccount []byte) (gcpclient.Interface, error)
}

// NewTerraformer is a mock for the NewTerraformer function.
type NewTerraformer interface {
	Do(config *rest.Config, account *internal.ServiceAccount, purpose, namespace, name string) (terraformer.Terraformer, error)
}

// CleanupKubernetesCloudArtifacts is a mock for the CleanupKubernetesCloudArtifacts function.
type CleanupKubernetesCloudArtifacts interface {
	Do(ctx context.Context, client gcpclient.Interface, projectID, network string) error
}

// TerraformerDefaultInitializer is a mock for the TerraformerDefaultInitializer function.
type TerraformerDefaultInitializer interface {
	Do(client client.Client, main, variables string, tfVars []byte) gardenterraformer.Initializer
}

// TerraformerInitializer is a mock for the TerraformerInitializer function.
type TerraformerInitializer interface {
	Do(config *gardenterraformer.InitializerConfig) error
}
