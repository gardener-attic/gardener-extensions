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

package operatingsystemconfig

import (
	"context"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"

	extensions1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type secretToOSCMapper struct {
	client     client.Client
	predicates []predicate.Predicate
}

func (m *secretToOSCMapper) Map(obj handler.MapObject) []reconcile.Request {
	if obj.Object == nil {
		return nil
	}

	secret, ok := obj.Object.(*corev1.Secret)
	if !ok {
		return nil
	}

	oscList := &extensions1alpha1.OperatingSystemConfigList{}
	if err := m.client.List(context.TODO(), client.InNamespace(secret.Namespace), oscList); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, osc := range oscList.Items {
		if !extensionscontroller.EvalGenericPredicate(&osc, m.predicates...) {
			continue
		}

		for _, file := range osc.Spec.Files {
			if secretRef := file.Content.SecretRef; secretRef != nil && secretRef.Name == secret.Name {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: osc.Namespace,
						Name:      osc.Name,
					},
				})
			}
		}
	}
	return requests
}

// SecretToOSCMapper returns a mapper that returns requests for OperatingSystemConfigs whose
// referenced secrets have been modified.
func SecretToOSCMapper(client client.Client, predicates []predicate.Predicate) handler.Mapper {
	return &secretToOSCMapper{client, predicates}
}
