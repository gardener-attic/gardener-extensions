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
	"context"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/certservice"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/rbac"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/lifecycle"

	configv1alpha1 "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config/v1alpha1"
	gcontroller "github.com/gardener/gardener-extensions/pkg/controller"
	controllercmd "github.com/gardener/gardener-extensions/pkg/controller/cmd"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/spf13/cobra"
)

// NewServiceControllerCommand creates a new command that is used to start the Certificate Service controller.
func NewServiceControllerCommand(ctx context.Context) *cobra.Command {
	options := NewOptions()

	cmd := &cobra.Command{
		Use:   "certificate-service-controller-manager",
		Short: "Certificate Service Controller manages components which provide certificate services.",

		Run: func(cmd *cobra.Command, args []string) {
			options.loadConfigOrDie()
			options.run(ctx)
		},
	}

	options.optionAggregator.AddFlags(cmd.Flags())

	return cmd
}

func (o *Options) run(ctx context.Context) error {
	mgr, err := manager.New(o.restOptions.Completed().Config, o.managerOptions.Completed().Options())
	if err != nil {
		controllercmd.LogErrAndExit(err, "Could not instantiate controller-manager")
		return err
	}

	if err := gcontroller.AddToScheme(mgr.GetScheme()); err != nil {
		controllercmd.LogErrAndExit(err, "Could not update manager scheme")
	}

	if err := configv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		controllercmd.LogErrAndExit(err, "Could not update manager scheme")
	}

	lifecycle.ResourceNamespace = o.certOptions.resourceNamespace
	lifecycle.ResourceName = o.certOptions.resourceName
	lifecycle.Sync = o.certOptions.lifecycleSync
	if err := lifecycle.AddToManager(mgr); err != nil {
		return err
	}

	certservice.ResourceNamespace = o.certOptions.resourceNamespace
	certservice.ResourceName = o.certOptions.resourceName
	o.controllerOptions.Completed().Apply(&certservice.Options)
	if err := certservice.AddToManager(mgr); err != nil {
		return err
	}

	rbac.ReconcileInterval = o.certOptions.rbacSync
	o.controllerOptions.Completed().Apply(&rbac.Options)
	if err := rbac.AddToManager(mgr); err != nil {
		return err
	}

	return mgr.Start(ctx.Done())
}
