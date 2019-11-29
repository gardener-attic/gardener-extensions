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

package predicate

import (
	"errors"

	machinesv1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"

	"github.com/gardener/gardener-extensions/pkg/controller"
	extensionsevent "github.com/gardener/gardener-extensions/pkg/event"
	extensionsinject "github.com/gardener/gardener-extensions/pkg/inject"

	"github.com/gardener/gardener/pkg/api/extensions"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// Log is the logger for predicates.
var Log logr.Logger = log.Log

// EvalGeneric returns true if all predicates match for the given object.
func EvalGeneric(obj runtime.Object, predicates ...predicate.Predicate) bool {
	e := extensionsevent.NewFromObject(obj)

	for _, p := range predicates {
		if !p.Generic(e) {
			return false
		}
	}

	return true
}

type shootNotFailedMapper struct {
	log logr.Logger
	extensionsinject.WithClient
	extensionsinject.WithContext
	extensionsinject.WithCache
}

func (s *shootNotFailedMapper) Map(e event.GenericEvent) bool {
	if e.Meta == nil {
		return false
	}

	// Wait for cache sync because of backing client cache.
	if !s.Cache.WaitForCacheSync(s.Context.Done()) {
		err := errors.New("failed to wait for caches to sync")
		s.log.Error(err, "Could not wait for Cache to sync", "predicate", "ShootNotFailed")
		return false
	}

	cluster, err := controller.GetCluster(s.Context, s.Client, e.Meta.GetNamespace())
	if err != nil {
		s.log.Error(err, "Could not retrieve corresponding cluster")
		return false
	}

	lastOperation := cluster.Shoot.Status.LastOperation
	return lastOperation != nil &&
		lastOperation.State != gardencorev1beta1.LastOperationStateFailed &&
		cluster.Shoot.Generation == cluster.Shoot.Status.ObservedGeneration
}

// ShootNotFailed is a predicate for failed shoots.
func ShootNotFailed() predicate.Predicate {
	return FromMapper(&shootNotFailedMapper{log: Log.WithName("shoot-not-failed")},
		CreateTrigger, UpdateNewTrigger, DeleteTrigger, GenericTrigger)
}

type or struct {
	predicates []predicate.Predicate
}

func (o *or) orRange(f func(predicate.Predicate) bool) bool {
	for _, p := range o.predicates {
		if f(p) {
			return true
		}
	}
	return false
}

// Create implements Predicate.
func (o *or) Create(event event.CreateEvent) bool {
	return o.orRange(func(p predicate.Predicate) bool { return p.Create(event) })
}

// Delete implements Predicate.
func (o *or) Delete(event event.DeleteEvent) bool {
	return o.orRange(func(p predicate.Predicate) bool { return p.Delete(event) })
}

// Update implements Predicate.
func (o *or) Update(event event.UpdateEvent) bool {
	return o.orRange(func(p predicate.Predicate) bool { return p.Update(event) })
}

// Generic implements Predicate.
func (o *or) Generic(event event.GenericEvent) bool {
	return o.orRange(func(p predicate.Predicate) bool { return p.Generic(event) })
}

// InjectFunc implements Injector.
func (o *or) InjectFunc(f inject.Func) error {
	for _, p := range o.predicates {
		if err := f(p); err != nil {
			return err
		}
	}
	return nil
}

// Or builds a logical OR gate of passed predicates.
func Or(predicates ...predicate.Predicate) predicate.Predicate {
	return &or{predicates}
}

// HasType filters the incoming OperatingSystemConfigs for ones that have the same type
// as the given type.
func HasType(typeName string) predicate.Predicate {
	return FromMapper(MapperFunc(func(e event.GenericEvent) bool {
		acc, err := extensions.Accessor(e.Object)
		if err != nil {
			return false
		}

		return acc.GetExtensionSpec().GetExtensionType() == typeName
	}), CreateTrigger, UpdateNewTrigger, DeleteTrigger, GenericTrigger)
}

// HasName returns a predicate that matches the given name of a resource.
func HasName(name string) predicate.Predicate {
	return FromMapper(MapperFunc(func(e event.GenericEvent) bool {
		return e.Meta.GetName() == name
	}), CreateTrigger, UpdateNewTrigger, DeleteTrigger, GenericTrigger)
}

// HasOperationAnnotation is a predicate for the operation annotation.
func HasOperationAnnotation() predicate.Predicate {
	return FromMapper(MapperFunc(func(e event.GenericEvent) bool {
		return e.Meta.GetAnnotations()[v1beta1constants.GardenerOperation] == v1beta1constants.GardenerOperationReconcile
	}), CreateTrigger, UpdateNewTrigger, GenericTrigger)
}

// LastOperationNotSuccessful is a predicate for unsuccessful last operations for creation events.
func LastOperationNotSuccessful() predicate.Predicate {
	operationNotSucceeded := func(obj runtime.Object) bool {
		acc, err := extensions.Accessor(obj)
		if err != nil {
			return false
		}

		lastOp := acc.GetExtensionStatus().GetLastOperation()
		return lastOp == nil ||
			lastOp.GetState() != gardencorev1beta1.LastOperationStateSucceeded
	}

	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return operationNotSucceeded(event.Object)
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return operationNotSucceeded(event.ObjectNew)
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return operationNotSucceeded(event.Object)
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return operationNotSucceeded(event.Object)
		},
	}
}

// IsDeleting is a predicate for objects having a deletion timestamp.
func IsDeleting() predicate.Predicate {
	return FromMapper(MapperFunc(func(e event.GenericEvent) bool {
		return e.Meta.GetDeletionTimestamp() != nil
	}), CreateTrigger, UpdateNewTrigger, GenericTrigger)
}

// AddTypePredicate returns a new slice which contains a type predicate and the given `predicates`.
func AddTypePredicate(extensionType string, predicates []predicate.Predicate) []predicate.Predicate {
	preds := make([]predicate.Predicate, 0, len(predicates)+1)
	preds = append(preds, HasType(extensionType))
	return append(preds, predicates...)
}

// MachineStatusHasChanged is a predicate deciding wether the status of a MCM's Machine has been changed.
func MachineStatusHasChanged() predicate.Predicate {
	statusHasChanged := func(oldObj runtime.Object, newObj runtime.Object) bool {
		oldMachine, ok := oldObj.(*machinesv1alpha1.Machine)
		if !ok {
			return false
		}
		newMachine, ok := newObj.(*machinesv1alpha1.Machine)
		if !ok {
			return false
		}
		oldStatus := oldMachine.Status
		newStatus := newMachine.Status

		//Check the Node
		if oldStatus.Node != newStatus.Node {
			return true
		}

		//Check the CurrentStatus
		if !equality.Semantic.DeepEqual(oldStatus.CurrentStatus, newStatus.CurrentStatus) {
			return true
		}

		if !equality.Semantic.DeepEqual(oldStatus.LastOperation, newStatus.LastOperation) {
			return true
		}

		//Check the Conditions
		if !equality.Semantic.DeepEqual(oldStatus.Conditions, newStatus.Conditions) {
			return true
		}

		return false
	}

	return statusChanged(statusHasChanged)
}

// MachineSetStatusHasChanged is a predicate deciding whether the status of a MCM's MachineSet has been changed.
func MachineSetStatusHasChanged() predicate.Predicate {
	statusHasChanged := func(oldObj runtime.Object, newObj runtime.Object) bool {
		oldMachineSet, ok := oldObj.(*machinesv1alpha1.MachineSet)
		if !ok {
			return false
		}
		newMachineSet, ok := newObj.(*machinesv1alpha1.MachineSet)
		if !ok {
			return false
		}
		oldStatus := oldMachineSet.Status
		newStatus := newMachineSet.Status

		//Check the The number of actual replicas
		if oldStatus.Replicas != newStatus.Replicas {
			return true
		}

		//Check the number of pods that have labels matching the labels of the pod template of the replicaset.
		if oldStatus.FullyLabeledReplicas != newStatus.FullyLabeledReplicas {
			return true
		}

		//Check the number of ready replicas for this replica set.
		if oldStatus.ReadyReplicas != newStatus.ReadyReplicas {
			return true
		}

		//Check the number of available replicas for this replica set.
		if oldStatus.AvailableReplicas != newStatus.AvailableReplicas {
			return true
		}

		//Check LastOperation
		if !equality.Semantic.DeepEqual(oldStatus.LastOperation, newStatus.LastOperation) {
			return true
		}

		//Check the MachineSetConditions
		if !equality.Semantic.DeepEqual(newStatus.Conditions, newStatus.Conditions) {
			return true
		}

		//Check MachineSummary
		if !equality.Semantic.DeepEqual(oldStatus.FailedMachines, newStatus.FailedMachines) {
			return true
		}

		return false
	}
	return statusChanged(statusHasChanged)
}

func statusChanged(statusHasChanged func(oldObj runtime.Object, newObj runtime.Object) bool) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			result := statusHasChanged(event.ObjectOld, event.ObjectNew)
			return result
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return false
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return true
		},
	}
}
