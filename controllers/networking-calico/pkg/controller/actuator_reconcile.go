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

package controller

import (
	"context"

	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/pkg/errors"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resourcemanager "github.com/gardener/gardener-resource-manager/pkg/manager"
	corev1 "k8s.io/api/core/v1"

	"github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/charts"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

const (
	calicoConfigSecretName = "extension-networking-calico-config"
)

func withLocalObjectRefs(refs ...string) []corev1.LocalObjectReference {
	var localObjectRefs []corev1.LocalObjectReference
	for _, ref := range refs {
		localObjectRefs = append(localObjectRefs, corev1.LocalObjectReference{Name: ref})
	}
	return localObjectRefs
}

func calicoSecret(cl client.Client, calicConfig []byte, namespace string) (*resourcemanager.Secret, []corev1.LocalObjectReference) {
	return resourcemanager.NewSecret(cl).
		WithKeyValues(map[string][]byte{charts.CalicoConfigKey: calicConfig}).
		WithNamespacedName(namespace, calicoConfigSecretName), withLocalObjectRefs(calicoConfigSecretName)
}

// Reconcile implements Network.Actuator.
func (a *actuator) Reconcile(ctx context.Context, network *extensionsv1alpha1.Network, cluster *extensionscontroller.Cluster) error {
	networkConfig, err := CalicoNetworkConfigFromNetworkResource(network)
	if err != nil {
		return err
	}

	// Create shoot chart renderer
	chartRenderer, err := a.chartRendererFactory.NewChartRendererForShoot(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return errors.Wrapf(err, "could not create chart renderer for shoot '%s'", network.Namespace)
	}

	calicoChart, err := charts.RenderCalicoChart(chartRenderer, network, networkConfig)
	if err != nil {
		return err
	}

	secret, secretRefs := calicoSecret(a.client, calicoChart, network.Namespace)
	err = secret.Reconcile(ctx)
	if err != nil {
		return err
	}

	if err := resourcemanager.NewManagedResource(a.client).
		WithNamespacedName(network.Namespace, calicoConfigSecretName).
		WithSecretRefs(secretRefs).
		WithInjectedLabels(map[string]string{common.ShootNoCleanup: "true"}).
		Reconcile(ctx); err != nil {
		return err
	}

	return a.updateProviderStatus(ctx, network, networkConfig)
}
