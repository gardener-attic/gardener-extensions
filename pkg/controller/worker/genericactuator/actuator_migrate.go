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

package genericactuator

import (
	"context"
	"fmt"
	"time"

	"github.com/gardener/gardener-extensions/pkg/controller"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/pkg/errors"
)

// Migrate remove all machine related resources (e.g. MachineDeployments, MachineClasses, MachineClassSecrets, MachineSets and Machines)
// without waiting for machine-cotroll-manager to do that. Before removal it ensures that the MCM is deleted.
func (a *genericActuator) Migrate(ctx context.Context, worker *extensionsv1alpha1.Worker, cluster *controller.Cluster) error {
	workerDelegate, err := a.delegateFactory.WorkerDelegate(ctx, worker, cluster)
	if err != nil {
		return errors.Wrapf(err, "could not instantiate actuator context")
	}

	// Make sure machine-controller-manager is deleted before deleting the machines.
	if err := a.deleteMachineControllerManager(ctx, worker); err != nil {
		return errors.Wrapf(err, "failed deleting machine-controller-manager")
	}

	// Delete all existing machines without deleting the VM corresponding to them.
	a.logger.Info("Shallow Deleting all machines", "worker", fmt.Sprintf("%s/%s", worker.Namespace, worker.Name))
	if err := a.shallowDeleteAllExistingMachines(ctx, worker.Namespace); err != nil {
		return errors.Wrapf(err, "Shallow Deletion of all machines failed")
	}
	// Delete all existing machineSets.
	a.logger.Info("Shallow Deleting all machineSets", "worker", fmt.Sprintf("%s/%s", worker.Namespace, worker.Name))
	if err := a.shallowDeleteAllExistingMachineSets(ctx, worker.Namespace); err != nil {
		return errors.Wrapf(err, "Shallow Deletion of all machineSets failed")
	}

	// Get the list of all existing machine deployments.
	existingMachineDeployments := &machinev1alpha1.MachineDeploymentList{}
	if err := a.client.List(ctx, existingMachineDeployments, client.InNamespace(worker.Namespace)); err != nil {
		return err
	}

	// Shallow delete all machine deployments.
	if err := a.shallowDeleteMachineDeployments(ctx, existingMachineDeployments, nil); err != nil {
		return errors.Wrapf(err, "cleaning up machine deployments failed")
	}

	// Delete all machine classes.
	if err := a.shallowDeleteMachineClasses(ctx, worker.Namespace, workerDelegate.MachineClassList(), nil); err != nil {
		return errors.Wrapf(err, "cleaning up machine classes failed")
	}

	// Delete all machine class secrets.
	if err := a.shallowDeleteMachineClassSecrets(ctx, worker.Namespace, nil); err != nil {
		return errors.Wrapf(err, "cleaning up machine class secrets failed")
	}

	// Wait until all machine resources have been properly deleted.
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := a.waitUntilMachineResourcesDeleted(timeoutCtx, worker, workerDelegate); err != nil {
		return gardencorev1beta1helper.DetermineError(fmt.Sprintf("Failed while waiting for all machine resources to be deleted: '%s'", err.Error()))
	}

	return nil
}
