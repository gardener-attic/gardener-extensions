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

	"github.com/gardener/gardener-extensions/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Secret struct {
	client client.Client

	keyValues map[string]string
	secret    corev1.Secret
}

func NewSecret(client client.Client) *Secret {
	return &Secret{
		client:    client,
		keyValues: make(map[string]string),
		secret: corev1.Secret{
			Type: corev1.SecretTypeOpaque,
		},
	}
}

func (s *Secret) WithNamespacedName(namespacedName types.NamespacedName) *Secret {
	s.secret.Namespace = namespacedName.Namespace
	s.secret.Name = namespacedName.Name

	return s
}

func (s *Secret) WithKeyValues(keyValues map[string][]byte) *Secret {
	s.secret.Data = keyValues

	return s
}

func (s *Secret) Reconcile(ctx context.Context) error {
	return controller.CreateOrUpdate(ctx, s.client, &s.secret, nil)
}

func (s *Secret) Delete(ctx context.Context) error {
	if err := s.client.Delete(ctx, &s.secret); err != nil && apierrors.IsNotFound(err) {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

type Secrets struct {
	client client.Client

	secrets []Secret
}

func NewSecrets(client client.Client) *Secrets {
	return &Secrets{
		client:  client,
		secrets: []Secret{},
	}
}

func (s *Secrets) WithSecretList(secrets []Secret) *Secrets {
	s.secrets = append(s.secrets, secrets...)

	return s
}

func (s *Secrets) WithSecret(secrets Secret) *Secrets {
	s.secrets = append(s.secrets, secrets)

	return s
}

func (s *Secrets) Reconcile(ctx context.Context) error {
	for _, secret := range s.secrets {
		if err := secret.Reconcile(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Secrets) Delete(ctx context.Context) error {
	for _, secret := range s.secrets {
		if err := secret.Delete(ctx); err != nil {
			return err
		}
	}
	return nil
}
