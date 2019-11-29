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

package worker

import (
	"context"
	"encoding/json"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
)

type genericStateActuator struct {
	logger logr.Logger

	client client.Client
}

// NewStateActuator creates a new Actuator that reconciles Worker's State subresource
// It provides a default implementation that allows easier integration of providers.
func NewStateActuator(logger logr.Logger) StateActuator {
	return &genericStateActuator{logger: logger.WithName("worker-state-actuator")}
}

func (a *genericStateActuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

// Reconcile update the Worker state with the latest.
func (a *genericStateActuator) Reconcile(ctx context.Context, worker *extensionsv1alpha1.Worker) error {
	copyOfWorker := worker.DeepCopy()
	if err := a.updateWorkerState(ctx, copyOfWorker); err != nil {
		return errors.Wrapf(err, "failed to update the state in worker status")
	}

	return nil
}

func (a *genericStateActuator) updateWorkerState(ctx context.Context, worker *extensionsv1alpha1.Worker) error {
	state, err := a.getWorkerState(ctx, worker)
	if err != nil {
		return err
	}

	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.client, worker, func() error {
		worker.Status.State = &runtime.RawExtension{Raw: state}
		return nil
	})
}

func (a *genericStateActuator) getWorkerState(ctx context.Context, worker *extensionsv1alpha1.Worker) ([]byte, error) {
	machineDeploymentNames, err := a.getExistingMachineDeploymentNames(ctx, worker)
	if err != nil {
		return nil, err
	}

	machineSets, err := a.getExistingMachineSetsMap(ctx, worker)
	if err != nil {
		return nil, err
	}

	machines, err := a.getExistingMachinesMap(ctx, worker)
	if err != nil {
		return nil, err
	}

	workerState := make(map[string]*MachineDeploymentState)
	for _, deploymentName := range machineDeploymentNames {
		machineDeploymentState := &MachineDeploymentState{}

		machineSet, ok := machineSets[deploymentName]
		if !ok {
			continue
		}
		addMachineSetToMachineDeploymentState(machineSet, machineDeploymentState)

		currentMachines := machines[machineSet.Name]
		if len(currentMachines) <= 0 {
			continue
		}

		for index := range currentMachines {
			addMachineToMachineDeploymentState(&currentMachines[index], machineDeploymentState)
		}

		workerState[deploymentName] = machineDeploymentState
	}

	return json.Marshal(workerState)
}

func (a *genericStateActuator) getExistingMachineDeploymentNames(ctx context.Context, worker *extensionsv1alpha1.Worker) ([]string, error) {
	existingMachineDeployments := &machinev1alpha1.MachineDeploymentList{}
	if err := a.client.List(ctx, existingMachineDeployments, client.InNamespace(worker.Namespace)); err != nil {
		return nil, err
	}

	var machineDeploymentNames []string
	for _, machineDeployment := range existingMachineDeployments.Items {
		machineDeploymentNames = append(machineDeploymentNames, machineDeployment.Name)
	}

	return machineDeploymentNames, nil
}

func (a *genericStateActuator) getExistingMachineSetsMap(ctx context.Context, worker *extensionsv1alpha1.Worker) (map[string]*machinev1alpha1.MachineSet, error) {
	existingMachineSets := &machinev1alpha1.MachineSetList{}
	if err := a.client.List(ctx, existingMachineSets, client.InNamespace(worker.Namespace)); err != nil {
		return nil, err
	}

	machineSets := make(map[string]*machinev1alpha1.MachineSet)
	for index, machineSet := range existingMachineSets.Items {
		for _, referant := range machineSet.OwnerReferences {
			if referant.Kind == "MachineDeployment" {
				machineSets[referant.Name] = &existingMachineSets.Items[index]
			}
		}
	}
	return machineSets, nil
}

func (a *genericStateActuator) getExistingMachinesMap(ctx context.Context, worker *extensionsv1alpha1.Worker) (map[string][]machinev1alpha1.Machine, error) {
	existingMachines := &machinev1alpha1.MachineList{}
	if err := a.client.List(ctx, existingMachines, client.InNamespace(worker.Namespace)); err != nil {
		return nil, err
	}

	machines := make(map[string][]machinev1alpha1.Machine)
	for index, machine := range existingMachines.Items {
		for _, referant := range machine.OwnerReferences {
			if referant.Kind == "MachineSet" {
				machines[referant.Name] = append(machines[referant.Name], existingMachines.Items[index])
			}
		}
	}
	return machines, nil
}

func addMachineSetToMachineDeploymentState(machineSet *machinev1alpha1.MachineSet, machineDeploymentState *MachineDeploymentState) {
	if machineSet == nil || machineDeploymentState == nil {
		return
	}

	//remove redundant data from the machine set
	machineSet.ObjectMeta = metav1.ObjectMeta{
		Name:        machineSet.Name,
		Namespace:   machineSet.Namespace,
		Annotations: machineSet.Annotations,
		Labels:      machineSet.Labels,
	}
	machineSet.OwnerReferences = nil
	machineSet.Status = machinev1alpha1.MachineSetStatus{}

	machineDeploymentState.MachineSet = &runtime.RawExtension{Object: machineSet}
}

func addMachineToMachineDeploymentState(machine *machinev1alpha1.Machine, machineDeploymentState *MachineDeploymentState) {
	if machine == nil || machineDeploymentState == nil {
		return
	}

	//remove redundant data from the machine
	machine.ObjectMeta = metav1.ObjectMeta{
		Name:        machine.Name,
		Namespace:   machine.Namespace,
		Annotations: machine.Annotations,
		Labels:      machine.Labels,
	}
	machine.OwnerReferences = nil
	machine.Status = machinev1alpha1.MachineStatus{
		Node: machine.Status.Node,
	}

	machineDeploymentState.Machines = append(machineDeploymentState.Machines, runtime.RawExtension{Object: machine})
}
