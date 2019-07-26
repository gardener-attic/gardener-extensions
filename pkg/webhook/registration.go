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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/secrets"
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

const (
	// A seed webhook is applied only to those shoot namespaces that have the correct Seed provider label.
	SeedKind = "seed"
	// A shoot webhook is applied only to those shoot namespaces that have the correct Shoot provider label.
	ShootKind = "shoot"
	// A backup webhook is applied only to those shoot namespaces that have the correct Backup provider label.
	BackupKind = "backup"

	// WebhookModeService is a constant for the webhook mode indicating that the controller is running inside of the Kubernetes cluster it
	// is serving.
	WebhookModeService = "service"
	// WebhookModeURL is a constant for the webhook mode indicating that the controller is running outside of the Kubernetes cluster it
	// is serving. If this is set then a URL is required for configuration.
	WebhookModeURL = "url"
)

// GenerateCertificates generates the certificates that are required for a webhook. It returns the ca bundle, and it
// stores the server certificate and key locally on the file system.
func GenerateCertificates(certDir, namespace, name, mode, url string) ([]byte, error) {
	var (
		serverKeyPath  = filepath.Join(certDir, "tls.key")
		serverCertPath = filepath.Join(certDir, "tls.crt")
	)

	caConfig := &secrets.CertificateSecretConfig{
		CommonName: "webhook-ca",
		CertType:   secrets.CACert,
	}

	caCert, err := caConfig.GenerateCertificate()
	if err != nil {
		return nil, err
	}

	var dnsNames []string
	switch mode {
	case WebhookModeURL:
		dnsNames = []string{
			url,
		}
	case WebhookModeService:
		dnsNames = []string{
			fmt.Sprintf("gardener-extension-%s", name),
			fmt.Sprintf("gardener-extension-%s.%s", name, namespace),
			fmt.Sprintf("gardener-extension-%s.%s.svc", name, namespace),
		}
	}

	serverConfig := &secrets.CertificateSecretConfig{
		CommonName: name,
		DNSNames:   dnsNames,
		CertType:   secrets.ServerCert,
		SigningCA:  caCert,
	}

	serverCert, err := serverConfig.GenerateCertificate()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(serverKeyPath, serverCert.PrivateKeyPEM, 0666); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(serverCertPath, serverCert.CertificatePEM, 0666); err != nil {
		return nil, err
	}

	return caCert.CertificatePEM, nil
}

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
	case SeedKind:
		key = gardencorev1alpha1.SeedProvider
	case ShootKind:
		key = gardencorev1alpha1.ShootProvider
	case BackupKind:
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
	case WebhookModeURL:
		url := fmt.Sprintf("https://%s:%d%s", url, port, path)
		clientConfig.URL = &url
	case WebhookModeService:
		clientConfig.Service = &admissionregistrationv1beta1.ServiceReference{
			Namespace: namespace,
			Name:      "gardener-extension-" + providerName,
			Path:      &path,
		}
	}

	return clientConfig
}
