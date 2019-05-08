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
	"fmt"
	"os"

	"github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/azure/install"
	azurecontroller "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/controller"
	azureinfrastructure "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/controller/infrastructure"
	"github.com/gardener/gardener-extensions/pkg/controller"
	controllercmd "github.com/gardener/gardener-extensions/pkg/controller/cmd"
	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"

	"github.com/spf13/cobra"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Name is the name of the Azure provider controller.
const Name = "provider-azure"

// NewControllerManagerCommand creates a new command for running a Azure provider controller.
func NewControllerManagerCommand(ctx context.Context) *cobra.Command {
	var (
		restOpts = &controllercmd.RESTOptions{}
		mgrOpts  = &controllercmd.ManagerOptions{
			LeaderElection:          true,
			LeaderElectionID:        controllercmd.LeaderElectionNameID(Name),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
		}

		infraCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}
		infraReconcileOpts = &infrastructure.ReconcilerOptions{
			IgnoreOperationAnnotation: true,
		}
		unprefixedInfraOpts = controllercmd.NewOptionAggregator(infraCtrlOpts, infraReconcileOpts)
		infraOpts           = controllercmd.PrefixOption("infrastructure-", &unprefixedInfraOpts)

		aggOption = controllercmd.NewOptionAggregator(restOpts, mgrOpts, infraOpts)
	)

	cmd := &cobra.Command{
		Use: fmt.Sprintf("%s-controller-manager", Name),

		Run: func(cmd *cobra.Command, args []string) {
			if err := aggOption.Complete(); err != nil {
				controllercmd.LogErrAndExit(err, "Error completing options")
			}

			mgr, err := manager.New(restOpts.Completed().Config, mgrOpts.Completed().Options())
			if err != nil {
				controllercmd.LogErrAndExit(err, "Could not instantiate manager")
			}

			if err := controller.AddToScheme(mgr.GetScheme()); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}

			if err := install.AddToScheme(mgr.GetScheme()); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}

			infraCtrlOpts.Completed().Apply(&azureinfrastructure.DefaultAddOptions.Controller)
			infraReconcileOpts.Completed().Apply(&azureinfrastructure.DefaultAddOptions.IgnoreOperationAnnotation)

			if err := azurecontroller.AddToManager(mgr, mgrOpts.Completed().Disabled); err != nil {
				controllercmd.LogErrAndExit(err, "Could not add controllers to manager")
			}

			if err := mgr.Start(ctx.Done()); err != nil {
				controllercmd.LogErrAndExit(err, "Error running manager")
			}
		},
	}

	aggOption.AddFlags(cmd.Flags())

	return cmd
}
