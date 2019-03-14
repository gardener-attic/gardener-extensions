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
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"google.golang.org/api/compute/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// KubernetesFirewallNamePrefix is the name prefix that Kubernetes related firewall rules have.
const KubernetesFirewallNamePrefix = "k8s"

// ListKubernetesFirewalls lists all firewalls that are in the given network and have the KubernetesFirewallNamePrefix.
func ListKubernetesFirewalls(ctx context.Context, client gcpclient.Interface, projectID, network string) ([]string, error) {
	var names []string
	err := client.Firewalls().List(projectID).Pages(ctx, func(list *compute.FirewallList) error {
		for _, firewall := range list.Items {
			if strings.HasSuffix(firewall.Network, network) && strings.HasPrefix(firewall.Name, KubernetesFirewallNamePrefix) {
				names = append(names, firewall.Name)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return names, nil
}

// DeleteFirewalls deletes the firewalls with the given names in the given project.
//
// If a deletion fails, it immediately returns the error of that deletion.
func DeleteFirewalls(ctx context.Context, client gcpclient.Interface, projectID string, firewalls []string) error {
	for _, firewall := range firewalls {
		if _, err := client.Firewalls().Delete(projectID, firewall).Context(ctx).Do(); err != nil {
			return err
		}
	}
	return nil
}

// CleanupKubernetesFirewalls lists all Kubernetes firewall rules and then deletes them one after another.
//
// If a deletion fails, this method returns immediately with the encountered error.
func CleanupKubernetesFirewalls(ctx context.Context, client gcpclient.Interface, projectID, network string) error {
	firewallNames, err := ListKubernetesFirewalls(ctx, client, projectID, network)
	if err != nil {
		return err
	}

	return DeleteFirewalls(ctx, client, projectID, firewallNames)
}

// GetServiceAccountFromInfrastructure retrieves the ServiceAccount from the Secret referenced in the given Infrastructure.
func GetServiceAccountFromInfrastructure(ctx context.Context, c client.Client, config *extensionsv1alpha1.Infrastructure) (*internal.ServiceAccount, error) {
	return internal.GetServiceAccount(ctx, c, config.Spec.SecretRef.Namespace, config.Spec.SecretRef.Name)
}
