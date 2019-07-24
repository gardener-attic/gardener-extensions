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
	"fmt"

	extensionwebhook "github.com/gardener/gardener-extensions/pkg/webhook"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// DisableFlag is the name of the command line flag to disable individual webhooks.
const DisableFlag = "disable-webhooks"

// NameToFactory binds a specific name to a webhook's factory function.
type NameToFactory struct {
	Name string
	Func func(manager.Manager) (*admission.Webhook, error)
}

// SwitchOptions are options to build an AddToManager function that filters the disabled webhooks.
type SwitchOptions struct {
	Disabled []string

	nameToWebhookFactory     map[string]func(manager.Manager) (*admission.Webhook, error)
	webhookFactoryAggregator extensionwebhook.FactoryAggregator
}

// Register registers the given NameToWebhookFuncs in the options.
func (w *SwitchOptions) Register(pairs ...NameToFactory) {
	for _, pair := range pairs {
		w.nameToWebhookFactory[pair.Name] = pair.Func
	}
}

// AddFlags implements Option.
func (w *SwitchOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&w.Disabled, DisableFlag, w.Disabled, "List of webhooks to disable")
}

// Complete implements Option.
func (w *SwitchOptions) Complete() error {
	disabled := sets.NewString()
	for _, disabledName := range w.Disabled {
		if _, ok := w.nameToWebhookFactory[disabledName]; !ok {
			return fmt.Errorf("cannot disable unknown webhook %q", disabledName)
		}
		disabled.Insert(disabledName)
	}

	for name, addToManager := range w.nameToWebhookFactory {
		if !disabled.Has(name) {
			w.webhookFactoryAggregator.Register(name, addToManager)
		}
	}
	return nil
}

// Completed returns the completed SwitchConfig. Call this only after successfully calling `Completed`.
func (w *SwitchOptions) Completed() *SwitchConfig {
	return &SwitchConfig{WebhooksFactory: w.webhookFactoryAggregator.Webhooks}
}

// SwitchConfig is the completed configuration of SwitchOptions.
type SwitchConfig struct {
	WebhooksFactory func(manager.Manager) (map[string]*admission.Webhook, error)
}

// Switch binds the given name to the given AddToManager function.
func Switch(name string, f func(manager.Manager) (*admission.Webhook, error)) NameToFactory {
	return NameToFactory{
		Name: name,
		Func: f,
	}
}

// NewSwitchOptions creates new SwitchOptions with the given initial pairs.
func NewSwitchOptions(pairs ...NameToFactory) *SwitchOptions {
	opts := SwitchOptions{nameToWebhookFactory: map[string]func(manager.Manager) (*admission.Webhook, error){}, webhookFactoryAggregator: extensionwebhook.FactoryAggregator{}}
	opts.Register(pairs...)
	return &opts
}

// AddToManagerOptions are options to create an `AddToManager` function from ServerOptions and SwitchOptions.
type AddToManagerOptions struct {
	serverName string
	Switch     SwitchOptions
}

// NewAddToManagerOptions creates new AddToManagerOptions with the given server name, server, and switch options.
func NewAddToManagerOptions(serverName string, switchOpts *SwitchOptions) *AddToManagerOptions {
	return &AddToManagerOptions{
		serverName: serverName,
		Switch:     *switchOpts,
	}
}

// AddFlags implements Option.
func (c *AddToManagerOptions) AddFlags(fs *pflag.FlagSet) {
	c.Switch.AddFlags(fs)
}

// Complete implements Option.
func (c *AddToManagerOptions) Complete() error {
	return c.Switch.Complete()
}

// Compoleted returns the completed AddToManagerConfig. Only call this if a previous call to `Complete` succeeded.
func (c *AddToManagerOptions) Completed() *AddToManagerConfig {
	return &AddToManagerConfig{
		serverName: c.serverName,
		Switch:     *c.Switch.Completed(),
	}
}

// AddToManagerConfig is a completed AddToManager configuration.
type AddToManagerConfig struct {
	serverName string
	Switch     SwitchConfig
}

// AddToManager instantiates all webhooks of this configuration. If there are any webhooks, it creates a
// webhook server, registers the webhooks and adds the server to the manager. Otherwise, it is a no-op.
func (c *AddToManagerConfig) AddToManager(mgr manager.Manager) error {
	webhooks, err := c.Switch.WebhooksFactory(mgr)
	if err != nil {
		return errors.Wrapf(err, "could not create webhooks")
	}

	webhookServer := mgr.GetWebhookServer()

	for name, wh := range webhooks {
		webhookServer.Register("/"+name, &webhook.Admission{Handler: wh})
	}

	return nil
}
