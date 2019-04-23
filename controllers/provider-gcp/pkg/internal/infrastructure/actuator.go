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
	"context"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal"
	gcpclient "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/client"
	"github.com/gardener/gardener-extensions/pkg/gardener/terraformer"
	gardenterraformer "github.com/gardener/gardener/pkg/operation/terraformer"

	gcpv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp/v1alpha1"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// NewGCPClientFromServiceAccount creates a new GCP client from the given service account.
	NewGCPClientFromServiceAccount = gcpclient.NewFromServiceAccount
	// NewTerraformer creates a new terraformer.
	NewTerraformer = internal.NewTerraformer
	// TerraformerDefaultInitializer creates the default initializer for a terraformer.
	TerraformerDefaultInitializer = gardenterraformer.DefaultInitializer
)

// Actuator is the infrastructure actuator.
type Actuator struct {
	Client        client.Client
	RESTConfig    *rest.Config
	ChartRenderer chartrenderer.Interface
}

// NewActuator creates a new infrastructure.Actuator.
func NewActuator() infrastructure.Actuator {
	return &Actuator{}
}

// InjectClient implements inject.Client.
func (a *Actuator) InjectClient(client client.Client) error {
	a.Client = client
	return nil
}

// InjectConfig implements inject.Config.
func (a *Actuator) InjectConfig(config *rest.Config) error {
	a.RESTConfig = config

	chartRenderer, err := chartrenderer.NewForConfig(config)
	if err != nil {
		return err
	}

	a.ChartRenderer = chartRenderer
	return nil
}

func (a *Actuator) updateProviderStatus(
	ctx context.Context,
	tf terraformer.Terraformer,
	infra *extensionsv1alpha1.Infrastructure,
	config *gcpv1alpha1.InfrastructureConfig,
) error {
	status, err := ComputeStatus(tf, config)
	if err != nil {
		return err
	}

	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.Client, infra, func() error {
		infra.Status.ProviderStatus = &runtime.RawExtension{Object: status}
		return nil
	})
}
