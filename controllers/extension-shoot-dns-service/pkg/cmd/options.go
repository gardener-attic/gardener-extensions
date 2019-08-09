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

	"github.com/gardener/gardener-extensions/controllers/extension-shoot-dns-service/pkg/controller"
	"github.com/gardener/gardener-extensions/controllers/extension-shoot-dns-service/pkg/controller/config"
	"github.com/gardener/gardener-extensions/pkg/controller/cmd"

	"github.com/spf13/pflag"
)

// DNSServiceOptions holds options related to the dns service.
type DNSServiceOptions struct {
	GardenID string
	SeedID   string
	DNSClass string
	config   *DNSServiceConfig
}

// AddFlags implements Flagger.AddFlags.
func (o *DNSServiceOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.GardenID, "garden-id", "", "ID of the current garden installation")
	fs.StringVar(&o.SeedID, "seed-id", "", "ID of the current cluster")
	fs.StringVar(&o.DNSClass, "dns-class", "garden", "DNS class used to filter DNS source resources in shoot clusters")
}

// Complete implements Completer.Complete.
func (o *DNSServiceOptions) Complete() error {
	if o.GardenID == "" {
		return fmt.Errorf("garden id must be specified")
	}
	if o.SeedID == "" {
		return fmt.Errorf("seed id must be specified")
	}
	o.config = &DNSServiceConfig{o.GardenID, o.SeedID, o.DNSClass}
	return nil
}

// Completed returns the decoded CertificatesServiceConfiguration instance. Only call this if `Complete` was successful.
func (o *DNSServiceOptions) Completed() *DNSServiceConfig {
	return o.config
}

// DNSServiceConfig contains configuration information about the dns service.
type DNSServiceConfig struct {
	GardenID string
	SeedID   string
	DNSClass string
}

// Apply applies the DNSServiceOptions to the passed ControllerOptions instance.
func (c *DNSServiceConfig) Apply(cfg *config.Config) {
	cfg.DNSServiceConfig.GardenID = c.GardenID
	cfg.DNSServiceConfig.SeedID = c.SeedID
	cfg.DNSServiceConfig.DNSClass = c.DNSClass
}

// SwitchOptions are the cmd.SwitchOptions for the provider controllers.
func ControllerSwitches() *cmd.SwitchOptions {
	return cmd.NewSwitchOptions(
		cmd.Switch(controller.Name, controller.AddToManager),
	)
}
