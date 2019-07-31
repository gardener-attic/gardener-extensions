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
	"context"
	"fmt"
	"strings"

	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Registers the given webhooks in the Kubernetes cluster targeted by the provided manager.
func RegisterWebhooks(ctx context.Context, mgr manager.Manager, namespace, providerName string, port int, mode, url string, caBundle []byte, webhooks []*Webhook) error {
	var (
		fail = admissionregistrationv1beta1.Fail

		mutatingWebhookConfiguration = &admissionregistrationv1beta1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gardener-extension-" + providerName,
			},
		}

		webhookToRegister []admissionregistrationv1beta1.Webhook
	)

	for _, webhook := range webhooks {
		namespaceSelector, err := buildSelector(webhook.Kind, webhook.Provider)
		if err != nil {
			return err
		}

		var rules []admissionregistrationv1beta1.RuleWithOperations
		for _, t := range webhook.Types {
			rule, err := buildRule(mgr, t)
			if err != nil {
				return err
			}
			rules = append(rules, *rule)
		}

		webhookToRegister = append(webhookToRegister, admissionregistrationv1beta1.Webhook{
			Name:              fmt.Sprintf("%s.%s.extensions.gardener.cloud", webhook.Name, strings.TrimPrefix(providerName, "provider-")),
			NamespaceSelector: namespaceSelector,
			FailurePolicy:     &fail,
			ClientConfig:      buildClientConfigFor(webhook, namespace, providerName, port, mode, url, caBundle),
			Rules:             rules,
		})
	}

	s := runtime.NewScheme()
	if err := scheme.AddToScheme(s); err != nil {
		return err
	}

	c, err := client.New(mgr.GetConfig(), client.Options{Scheme: s})
	if err != nil {
		return err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, c, mutatingWebhookConfiguration, func() error {
		mutatingWebhookConfiguration.Webhooks = webhookToRegister
		return nil
	})
	return err
}

// buildSelector creates and returns a LabelSelector for the given webhook kind and provider.
func buildSelector(kind, provider string) (*metav1.LabelSelector, error) {
	// Determine label selector key from the kind
	var key string
	switch kind {
	case controlplane.KindSeed:
		key = gardencorev1alpha1.SeedProvider
	case controlplane.KindShoot:
		key = gardencorev1alpha1.ShootProvider
	case controlplane.KindBackup:
		key = gardencorev1alpha1.BackupProvider
	default:
		return nil, fmt.Errorf("invalid webhook kind '%s'", kind)
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

func buildClientConfigFor(webhook *Webhook, namespace, providerName string, port int, mode, url string, caBundle []byte) admissionregistrationv1beta1.WebhookClientConfig {
	path := "/" + webhook.Path

	clientConfig := admissionregistrationv1beta1.WebhookClientConfig{
		CABundle: caBundle,
	}

	switch mode {
	case ModeURL:
		url := fmt.Sprintf("https://%s:%d%s", url, port, path)
		clientConfig.URL = &url
	case ModeService:
		clientConfig.Service = &admissionregistrationv1beta1.ServiceReference{
			Namespace: namespace,
			Name:      "gardener-extension-" + providerName,
			Path:      &path,
		}
	}

	return clientConfig
}
