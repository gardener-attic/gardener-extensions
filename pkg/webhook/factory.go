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

package webhook

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Webhook kinds.
const (
	// A seed webhook is applied only to those shoot namespaces that have the correct Seed provider label.
	SeedKind = "seed"
	// A shoot webhook is applied only to those shoot namespaces that have the correct Shoot provider label.
	ShootKind = "shoot"
	// A backup webhook is applied only to those shoot namespaces that have the correct Backup provider label.
	BackupKind = "backup"
)

// FactoryAggregator aggregates various Factory functions.
type FactoryAggregator map[string]func(manager.Manager) (*admission.Webhook, error)

// NewFactoryAggregator creates a new FactoryAggregator and registers the given functions.
func NewFactoryAggregator(m map[string]func(manager.Manager) (*admission.Webhook, error)) FactoryAggregator {
	builder := FactoryAggregator{}

	for name, f := range m {
		builder.Register(name, f)
	}

	return builder
}

// Register registers the given functions in this builder.
func (a *FactoryAggregator) Register(name string, f func(manager.Manager) (*admission.Webhook, error)) {
	(*a)[name] = f
}

// Webhooks calls all factories with the given managers and returns all created webhooks.
// As soon as there is an error creating a webhook, the error is returned immediately.
func (a *FactoryAggregator) Webhooks(mgr manager.Manager) (map[string]*admission.Webhook, error) {
	webhooks := make(map[string]*admission.Webhook, len(*a))

	for name, f := range *a {
		wh, err := f(mgr)
		if err != nil {
			return nil, err
		}

		webhooks[name] = wh
	}

	return webhooks, nil
}
