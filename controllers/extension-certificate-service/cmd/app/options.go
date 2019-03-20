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

package app

import (
	"os"
	"time"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/certservice"
	controllercmd "github.com/gardener/gardener-extensions/pkg/controller/cmd"
	"github.com/spf13/pflag"
)

// Options holds configuration passed to the Certificate Service controller.
type Options struct {
	certOptions       *CertificateServiceOptions
	restOptions       *controllercmd.RESTOptions
	managerOptions    *controllercmd.ManagerOptions
	controllerOptions *controllercmd.ControllerOptions
	optionAggregator  controllercmd.OptionAggregator
}

// NewOptions creates a new Options instance.
func NewOptions() *Options {
	options := &Options{
		certOptions: &CertificateServiceOptions{},
		restOptions: &controllercmd.RESTOptions{},
		managerOptions: &controllercmd.ManagerOptions{
			// These are default values.
			LeaderElection:          true,
			LeaderElectionID:        controllercmd.LeaderElectionNameID(certservice.ControllerName),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
		},
		controllerOptions: &controllercmd.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
	}

	options.optionAggregator = controllercmd.NewOptionAggregator(
		options.restOptions,
		options.managerOptions,
		options.controllerOptions,
		options.certOptions,
	)

	return options
}

func (o *Options) loadConfigOrDie() {
	if err := o.optionAggregator.Complete(); err != nil {
		controllercmd.LogErrAndExit(err, "Error completing options")
	}
}

// CertificateServiceOptions holds options related to the certificate service.
type CertificateServiceOptions struct {
	resourceNamespace string
	resourceName      string
	lifecycleSync     time.Duration
	rbacSync          time.Duration
}

// AddFlags implements Flagger.AddFlags.
func (o *CertificateServiceOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.resourceNamespace, "resource-namespace", "kube-system", "Namespace used to fetch configuration information and create resources for Cert-Manager")
	fs.StringVar(&o.resourceName, "resource-name", "certificate-service", "Name of the CertServiceConfig")
	fs.DurationVar(&o.lifecycleSync, "lifecycle-sync", 1*time.Hour, "Interval in which the configuration for this certificate service is reconciled")
	fs.DurationVar(&o.rbacSync, "rbac-sync", 5*time.Minute, "Interval in which the RBAC manifests for target clusters are re-applied")
}

// Complete implements Completer.Complete.
func (o *CertificateServiceOptions) Complete() error {
	return nil
}
