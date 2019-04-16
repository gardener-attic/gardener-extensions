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
	"github.com/gardener/gardener-extensions/pkg/webhook/cmd"
	"github.com/pkg/errors"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
)

const (
	// NameSuffix is a common suffix for all webhook names.
	NameSuffix = "extensions.gardener.cloud"

	// SeedProviderLabel is a label on shoot namespaces in the seed cluster that identifies the Seed provider.
	// TODO Move this constant to gardener/gardener
	SeedProviderLabel = "seed.gardener.cloud/provider"
	// ShootProviderLabel is a label on shoot namespaces in the seed cluster that identifies the Shoot provider.
	// TODO Move this constant to gardener/gardener
	ShootProviderLabel = "shoot.gardener.cloud/provider"
)

// Kind is a type for webhook kinds.
type Kind string

// Webhook kinds.
const (
	// A seed webhook is applied only to those shoot namespaces that have the correct Seed provider label.
	SeedKind Kind = "seed"
	// A shoot webhook is applied only to those shoot namespaces that have the correct Shoot provider label.
	ShootKind Kind = "shoot"
)

// AddToManagerBuilder aggregates various AddToManager functions.
type AddToManagerBuilder []func(manager.Manager) (webhook.Webhook, error)

// NewAddToManagerBuilder creates a new AddToManagerBuilder and registers the given functions.
func NewAddToManagerBuilder(funcs ...func(manager.Manager) (webhook.Webhook, error)) AddToManagerBuilder {
	var builder AddToManagerBuilder
	builder.Register(funcs...)
	return builder
}

// Register registers the given functions in this builder.
func (a *AddToManagerBuilder) Register(funcs ...func(manager.Manager) (webhook.Webhook, error)) {
	*a = append(*a, funcs...)
}

// AddToManager creates a webhook server adds all webhooks created by the AddToManager-functions of this builder to it.
// It exits on the first error and returns it.
func (a *AddToManagerBuilder) AddToManager(mgr manager.Manager, cfg *cmd.WebhookServerConfig) error {
	logger := log.Log.WithName("webhook-server")

	var whs []webhook.Webhook
	for _, f := range *a {
		wh, err := f(mgr)
		if err != nil {
			return err
		}
		whs = append(whs, wh)
	}

	if len(whs) > 0 {
		// Create webhook server
		logger.Info("Creating webhook server", "cfg", cfg)
		svr, err := newServer(mgr, cfg)
		if err != nil {
			return err
		}

		// Register webhooks with server
		logger.Info("Registering webhooks")
		err = svr.Register(whs...)
		if err != nil {
			return errors.Wrap(err, "could not register webhooks with server")
		}
	}
	return nil
}

func newServer(mgr manager.Manager, cfg *cmd.WebhookServerConfig) (*webhook.Server, error) {
	// Create webhook server
	disableConfigInstaller := false
	svr, err := webhook.NewServer("webhook-server", mgr, webhook.ServerOptions{
		Port:                          cfg.Port,
		CertDir:                       cfg.CertDir,
		DisableWebhookConfigInstaller: &disableConfigInstaller,
		BootstrapOptions:              cfg.BootstrapOptions,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not create webhook server")
	}
	return svr, nil
}

// NewWebhook creates a new mutating webhook for create and update operations
// with the given kind, provider, and name, applicable to objects of all given types,
// executing the given handler, and bound to the given manager.
func NewWebhook(mgr manager.Manager, kind Kind, provider, name string, types []runtime.Object, handler admission.Handler) (*admission.Webhook, error) {
	// Build namespace selector from the webhook kind and provider
	namespaceSelector, err := buildSelector(kind, provider)
	if err != nil {
		return nil, err
	}

	// Build rules for all object types
	var rules []admissionregistrationv1beta1.RuleWithOperations
	for _, t := range types {
		rule, err := buildRule(mgr, t)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *rule)
	}

	// Build webhook
	return builder.NewWebhookBuilder().
		Name(name + "." + provider + "." + NameSuffix).
		Path("/" + name).
		Mutating().
		FailurePolicy(admissionregistrationv1beta1.Fail).
		NamespaceSelector(namespaceSelector).
		Rules(rules...).
		Handlers(handler).
		WithManager(mgr).
		Build()
}

// buildSelector creates and returns a LabelSelector for the given webhook kind and provider.
func buildSelector(kind Kind, provider string) (*metav1.LabelSelector, error) {
	// Determine label selector key from the kind
	var key string
	switch kind {
	case SeedKind:
		key = SeedProviderLabel
	case ShootKind:
		key = ShootProviderLabel
	default:
		return nil, errors.Errorf("invalid webhook kind '%s'", kind)
	}

	// Create and return LabelSelector
	return &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{Key: key, Operator: metav1.LabelSelectorOpIn, Values: []string{provider}},
		},
	}, nil
}

// buildRule creates and returns a RuleWithOperations for the given object type.
func buildRule(mgr manager.Manager, t runtime.Object) (*admissionregistrationv1beta1.RuleWithOperations, error) {
	// Get GVK from the type
	gvk, err := apiutil.GVKForObject(t, mgr.GetScheme())
	if err != nil {
		return nil, errors.Wrapf(err, "could not get GroupVersionKind from object %v", t)
	}

	// Get REST mapping from GVK
	mapping, err := mgr.GetRESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get REST mapping from GroupVersionKind '%s'", gvk.String())
	}

	// Create and return RuleWithOperations
	return &admissionregistrationv1beta1.RuleWithOperations{
		Operations: []admissionregistrationv1beta1.OperationType{
			admissionregistrationv1beta1.Create,
			admissionregistrationv1beta1.Update,
		},
		Rule: admissionregistrationv1beta1.Rule{
			APIGroups:   []string{gvk.Group},
			APIVersions: []string{gvk.Version},
			Resources:   []string{mapping.Resource.Resource},
		},
	}, nil
}
