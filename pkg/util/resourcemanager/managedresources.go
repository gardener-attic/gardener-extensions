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

package resourcemanager

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/gardener/gardener-extensions/pkg/controller"
	resourcemanagerv1alpha1 "github.com/gardener/gardener-resource-manager/pkg/apis/resources/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ManagedResource struct {
	client   client.Client
	resource resourcemanagerv1alpha1.ManagedResource
}

func NewManagedResource(client client.Client) *ManagedResource {
	return &ManagedResource{
		client: client,
		resource: resourcemanagerv1alpha1.ManagedResource{
			Spec: resourcemanagerv1alpha1.ManagedResourceSpec{
				SecretRefs:   []corev1.LocalObjectReference{},
				InjectLabels: map[string]string{},
			},
		},
	}
}

func (m *ManagedResource) WithNamespacedName(namespacedName types.NamespacedName) *ManagedResource {
	m.resource.Namespace = namespacedName.Namespace
	m.resource.Name = namespacedName.Name

	return m
}

func (m *ManagedResource) WithSecretRef(secretRefName string) *ManagedResource {
	m.resource.Spec.SecretRefs = append(m.resource.Spec.SecretRefs, corev1.LocalObjectReference{Name: secretRefName})

	return m
}

func (m *ManagedResource) WithSecretRefs(secretRefs []corev1.LocalObjectReference) *ManagedResource {
	m.resource.Spec.SecretRefs = append(m.resource.Spec.SecretRefs, secretRefs...)

	return m
}

func (m *ManagedResource) WithInjectedLabels(labelsToInject map[string]string) *ManagedResource {
	m.resource.Spec.InjectLabels = labelsToInject

	return m
}

func (m *ManagedResource) Reconcile(ctx context.Context) error {
	return controller.CreateOrUpdate(ctx, m.client, &m.resource, nil)
}

func (m *ManagedResource) Delete(ctx context.Context) error {
	if err := m.client.Delete(ctx, &m.resource); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}
