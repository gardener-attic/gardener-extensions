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
	gcpv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal"
	gcpclient "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/client"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/infrastructure"
	"github.com/gardener/gardener-extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/terraformer"
	"github.com/gardener/gardener/pkg/utils/flow"
	"time"
)

func (a *actuator) cleanupKubernetesFirewallRules(
	ctx context.Context,
	config *gcpv1alpha1.InfrastructureConfig,
	client gcpclient.Interface,
	tf *terraformer.Terraformer,
	account *internal.ServiceAccount,
) error {
	state, err := infrastructure.ExtractTerraformState(tf, config)
	if err != nil {
		if terraformer.IsVariablesNotFoundError(err) {
			return nil
		}
		return err
	}

	return infrastructure.CleanupKubernetesFirewalls(ctx, client, account.ProjectID, state.VPCName)
}

// Delete implements infrastructure.Actuator.
func (a *actuator) Delete(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *controller.Cluster) error {
	config, err := internal.InfrastructureConfigFromInfrastructure(infra)
	if err != nil {
		return err
	}

	serviceAccount, err := internal.GetServiceAccount(ctx, a.client, infra.Spec.SecretRef.Namespace, infra.Spec.SecretRef.Name)
	if err != nil {
		return err
	}

	gcpClient, err := gcpclient.NewFromServiceAccount(ctx, serviceAccount.Raw)
	if err != nil {
		return err
	}

	tf, err := internal.NewTerraformer(a.restConfig, serviceAccount, infrastructure.TerraformerPurpose, infra.Namespace, infra.Name)
	if err != nil {
		return err
	}

	configExists, err := tf.ConfigExists()
	if err != nil {
		return err
	}

	var (
		g                              = flow.NewGraph("GCP infrastructure destruction")
		destroyKubernetesFirewallRules = g.Add(flow.Task{
			Name: "Destroying Kubernetes firewall rules",
			Fn: flow.TaskFn(func(ctx context.Context) error {
				return a.cleanupKubernetesFirewallRules(ctx, config, gcpClient, tf, serviceAccount)
			}).
				RetryUntilTimeout(10*time.Second, 5*time.Minute).
				DoIf(configExists),
		})

		_ = g.Add(flow.Task{
			Name:         "Destroying Shoot infrastructure",
			Fn:           flow.SimpleTaskFn(tf.Destroy),
			Dependencies: flow.NewTaskIDs(destroyKubernetesFirewallRules),
		})

		f = g.Compile()
	)

	if err := f.Run(flow.Opts{Context: ctx}); err != nil {
		return flow.Causes(err)
	}
	return nil
}
