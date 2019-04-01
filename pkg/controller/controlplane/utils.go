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

package controlplane

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

// DNSNamesForService returns the possible DNS names for a service with the given name and namespace
func DNSNamesForService(name, namespace string) []string {
	return []string{
		name,
		fmt.Sprintf("%s.%s", name, namespace),
		fmt.Sprintf("%s.%s.svc", name, namespace),
		fmt.Sprintf("%s.%s.svc.%s", name, namespace, gardenv1beta1.DefaultDomain),
	}
}

// ComputeChecksums computes and returns the checksums for the given secrets and configmaps,
// as well as the secrets and configmaps with the given names that are fetched from the cluster.
func ComputeChecksums(
	ctx context.Context,
	c client.Client,
	secrets map[string]*corev1.Secret,
	configMaps map[string]*corev1.ConfigMap,
	secretNames, configMapNames []string,
	namespace string,
) (map[string]string, error) {
	// Get cluster secrets
	clusterSecrets := make(map[string]*corev1.Secret)
	for _, name := range secretNames {
		secret := &corev1.Secret{}
		if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, secret); err != nil {
			return nil, errors.Wrapf(err, "could not get secret '%s/%s'", namespace, name)
		}
		clusterSecrets[name] = secret
	}

	// Get cluster configmaps
	clusterConfigMaps := make(map[string]*corev1.ConfigMap)
	for _, name := range configMapNames {
		cm := &corev1.ConfigMap{}
		if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, cm); err != nil {
			return nil, errors.Wrapf(err, "could not get configmap '%s/%s'", namespace, name)
		}
		clusterConfigMaps[name] = cm
	}

	// Compute checksums
	return computeChecksums(mergeSecretMaps(secrets, clusterSecrets), mergeConfigMapMaps(configMaps, clusterConfigMaps)), nil
}

// mergeSecretMaps merges the 2 given secret maps.
func mergeSecretMaps(a, b map[string]*corev1.Secret) map[string]*corev1.Secret {
	x := make(map[string]*corev1.Secret)
	for _, m := range []map[string]*corev1.Secret{a, b} {
		for k, v := range m {
			x[k] = v
		}
	}
	return x
}

// mergeConfigMapMaps merges the 2 given configmap maps.
func mergeConfigMapMaps(a, b map[string]*corev1.ConfigMap) map[string]*corev1.ConfigMap {
	x := make(map[string]*corev1.ConfigMap)
	for _, m := range []map[string]*corev1.ConfigMap{a, b} {
		for k, v := range m {
			x[k] = v
		}
	}
	return x
}

// computeChecksums computes and returns SAH256 checksums for the given secrets and configmaps.
func computeChecksums(secrets map[string]*corev1.Secret, cms map[string]*corev1.ConfigMap) map[string]string {
	checksums := make(map[string]string, len(secrets)+len(cms))
	for name, secret := range secrets {
		checksums[name] = computeChecksum(secret.Data)
	}
	for name, cm := range cms {
		checksums[name] = computeChecksum(cm.Data)
	}
	return checksums
}

func computeChecksum(data interface{}) string {
	jsonString, _ := json.Marshal(data)
	return utils.ComputeSHA256Hex(jsonString)
}
