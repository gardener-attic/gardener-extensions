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

package cmd

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	// PortFlag is the name of the command line flag to specify the webhook server port.
	PortFlag = "webhook-server-port"
	// CertDirFlag is the name of the command line flag to specify the directory that contains the webhook server key and certificate.
	CertDirFlag = "webhook-server-cert-dir"
	// ModeFlag is the name of the command line flag to specify the webhook config mode, either 'service' or 'url'.
	ModeFlag = "webhook-config-mode"
	// NameFlag is the name of the command line flag to specify the webhook config name.
	NameFlag = "webhook-config-name"
	// NamespaceFlag is the name of the command line flag to specify the webhook config namespace for 'service' mode.
	NamespaceFlag = "webhook-config-namespace"
	// ServiceSelectorsFlag is the name of the command line flag to specify the webhook config service selectors as JSON for 'service' mode.
	ServiceSelectorsFlag = "webhook-config-service-selectors"
	// HostFlag is the name of the command line flag to specify the webhook config host for 'url' mode.
	HostFlag = "webhook-config-host"
)

// Webhook config modes
const (
	ServiceMode = "service"
	URLMode     = "url"
)

// WebhookServerOptions are command line options that can be set for WebhookServerConfig.
type WebhookServerOptions struct {
	// Port is the webhook server port.
	Port int32
	// CertDir is the directory that contains the webhook server key and certificate.
	CertDir string
	// Mode is the webhook config mode, either 'service' or 'url'
	Mode string
	// Name is the webhook config name.
	Name string
	// Namespace is the webhook config namespace for 'service' mode.
	Namespace string
	// ServiceSelectors is the webhook config service selectors as JSON for 'service' mode.
	ServiceSelectors string
	// Host is the webhook config host for 'url' mode.
	Host string

	config *WebhookServerConfig
}

// WebhookServerConfig is a completed webhook server configuration.
type WebhookServerConfig struct {
	// Port is the webhook server port.
	Port int32
	// CertDir is the directory that contains the webhook server key and certificate.
	CertDir string
	// BootstrapOptions contains the options for bootstrapping the webhook server.
	BootstrapOptions *webhook.BootstrapOptions
}

// Complete implements Completer.Complete.
func (w *WebhookServerOptions) Complete() error {
	bootstrapOptions, err := w.buildBootstrapOptions()
	if err != nil {
		return err
	}

	w.config = &WebhookServerConfig{
		Port:             w.Port,
		CertDir:          w.CertDir,
		BootstrapOptions: bootstrapOptions,
	}
	return nil
}

// Completed returns the completed WebhookServerConfig. Only call this if `Complete` was successful.
func (w *WebhookServerOptions) Completed() *WebhookServerConfig {
	return w.config
}

// AddFlags implements Flagger.AddFlags.
func (w *WebhookServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.Int32Var(&w.Port, PortFlag, w.Port, "The webhook server port.")
	fs.StringVar(&w.CertDir, CertDirFlag, w.CertDir, "The directory that contains the webhook server key and certificate.")
	fs.StringVar(&w.Mode, ModeFlag, w.Mode, "The webhook config mode, either 'service' or 'url'.")
	fs.StringVar(&w.Name, NameFlag, w.Name, "The webhook config name.")
	fs.StringVar(&w.Namespace, NamespaceFlag, w.Namespace, "The webhook config namespace for 'service' mode.")
	fs.StringVar(&w.ServiceSelectors, ServiceSelectorsFlag, w.ServiceSelectors, "The webhook config service selectors as JSON for 'service' mode.")
	fs.StringVar(&w.Host, HostFlag, w.Host, "The webhook config host for 'url' mode.")
}

func (w *WebhookServerOptions) buildBootstrapOptions() (*webhook.BootstrapOptions, error) {
	switch w.Mode {
	case ServiceMode:
		serviceSelectors := make(map[string]string)
		if err := json.Unmarshal([]byte(w.ServiceSelectors), &serviceSelectors); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal webhook config service selectors from JSON")
		}

		return &webhook.BootstrapOptions{
			MutatingWebhookConfigName: w.Name,
			Secret:                    &types.NamespacedName{Namespace: w.Namespace, Name: w.Name},
			Service: &webhook.Service{
				Name:      w.Name,
				Namespace: w.Namespace,
				Selectors: serviceSelectors,
			},
		}, nil

	case URLMode:
		return &webhook.BootstrapOptions{
			MutatingWebhookConfigName: w.Name,
			Host:                      &w.Host,
		}, nil

	default:
		return nil, errors.Errorf("invalid webhook config mode '%s'", w.Mode)
	}
}
