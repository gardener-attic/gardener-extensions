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

	azurev1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/azure/v1alpha1"
	infrainternal "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/internal/infrastructure"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"
	"github.com/gardener/gardener-extensions/pkg/terraformer"

	"github.com/go-logr/logr"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type actuator struct {
	logger        logr.Logger
	client        client.Client
	restConfig    *rest.Config
	chartRenderer chartrenderer.Interface
}

// NewActuator creates a new infrastructure.Actuator.
func NewActuator() infrastructure.Actuator {
	return &actuator{
		logger: log.Log.WithName("infrastructure-actuator"),
	}
}

// InjectClient implements inject.Client.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

// InjectConfig implements inject.Config.
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.restConfig = config

	chartRenderer, err := chartrenderer.NewForConfig(config)
	if err != nil {
		return err
	}

	a.chartRenderer = chartRenderer
	return nil
}

func (a *actuator) updateProviderStatus(
	ctx context.Context,
	tf *terraformer.Terraformer,
	infra *extensionsv1alpha1.Infrastructure,
	config *azurev1alpha1.InfrastructureConfig,
) error {
	status, err := infrainternal.ComputeStatus(tf, config)
	if err != nil {
		return err
	}

	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.client, infra, func() error {
		infra.Status.ProviderStatus = &runtime.RawExtension{Object: status}
		return nil
	})
}
