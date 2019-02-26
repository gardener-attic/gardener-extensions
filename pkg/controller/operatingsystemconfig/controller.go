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

package operatingsystemconfig

import (
	"context"

	controllerutil "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/util"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	// FinalizerName is the name of the finalizer written by this controller.
	FinalizerName = "extensions.gardener.cloud/operatingsystemconfigs"

	name = "operatingsystemconfig-controller"
)

// AddArgs are arguments for adding an operatingsystemconfig controller to a manager.
type AddArgs struct {
	// Actuator is an operatingsystemconfig actuator.
	Actuator Actuator
	// Type is the operatingsystemconfig type the actuator supports.
	Type string
	// ControllerOptions are the controller options used for creating a controller.
	// The options.Reconciler is always overridden with a reconciler created from the
	// given actuator.
	ControllerOptions controller.Options
	// Predicates are the predicates to use.
	// If unset, GenerationChangedPredicate will be used.
	Predicates []predicate.Predicate
}

// Add adds an operatingsystemconfig controller to the given manager using the given AddArgs.
func Add(mgr manager.Manager, args AddArgs) error {
	args.ControllerOptions.Reconciler = NewReconciler(args.Actuator)
	return add(mgr, args.Type, args.ControllerOptions, args.Predicates)
}

func add(mgr manager.Manager, typeName string, options controller.Options, predicates []predicate.Predicate) error {
	ctrl, err := controller.New("operatingsystemconfig-controller", mgr, options)
	if err != nil {
		return err
	}

	if predicates == nil {
		predicates = append(predicates, GenerationChangedPredicate())
	}
	predicates = append(predicates, TypePredicate(typeName))

	if err := ctrl.Watch(&source.Kind{Type: &extensionsv1alpha1.OperatingSystemConfig{}}, &handler.EnqueueRequestForObject{}, predicates...); err != nil {
		return err
	}

	if err := ctrl.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: SecretToOSCMapper(mgr.GetClient(), typeName)}); err != nil {
		return err
	}

	return nil
}

// reconciler reconciles OperatingSystemConfig resources of Gardener's
// `extensions.gardener.cloud` API group.
type reconciler struct {
	logger   logr.Logger
	actuator Actuator

	ctx    context.Context
	client client.Client
}

var _ reconcile.Reconciler = &reconciler{}

// NewReconciler creates a new reconcile.Reconciler that reconciles
// OperatingSystemConfig resources of Gardener's `extensions.gardener.cloud` API group.
func NewReconciler(actuator Actuator) reconcile.Reconciler {
	logger := log.Log.WithName(name)
	return &reconciler{logger: logger, actuator: actuator}
}

// InjectFunc enables dependency injection into the actuator.
func (r *reconciler) InjectFunc(f inject.Func) error {
	return f(r.actuator)
}

// InjectClient injects the controller runtime client into the reconciler.
func (r *reconciler) InjectClient(client client.Client) error {
	r.client = client
	return nil
}

// InjectStopChannel is an implementation for getting the respective stop channel managed by the controller-runtime.
func (r *reconciler) InjectStopChannel(stopCh <-chan struct{}) error {
	r.ctx = util.ContextFromStopChannel(stopCh)
	return nil
}

// Reconcile is the reconciler function that gets executed in case there are new events for the `OperatingSystemConfig`
// resources.
func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	osc := &extensionsv1alpha1.OperatingSystemConfig{}
	if err := r.client.Get(r.ctx, request.NamespacedName, osc); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		r.logger.Error(err, "Could not fetch OperatingSystemConfig")
		return reconcile.Result{}, err
	}

	if osc.DeletionTimestamp != nil {
		return r.delete(r.ctx, osc)
	}
	return r.reconcile(r.ctx, osc)
}

func (r *reconciler) reconcile(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig) (reconcile.Result, error) {
	if err := controllerutil.EnsureFinalizer(ctx, r.client, FinalizerName, osc); err != nil {
		return reconcile.Result{}, err
	}

	exist, err := r.actuator.Exists(ctx, osc)
	if err != nil {
		return reconcile.Result{}, err
	}

	if exist {
		r.logger.Info("Reconciling operating system config triggers idempotent update.", "osc", osc.Name)
		if err := r.actuator.Update(ctx, osc); err != nil {
			return controllerutil.ReconcileErr(err)
		}
		return reconcile.Result{}, nil
	}

	r.logger.Info("Reconciling operating system config triggers idempotent create.", "osc", osc.Name)
	if err := r.actuator.Create(ctx, osc); err != nil {
		r.logger.Error(err, "Unable to create operating system config", "osc", osc.Name)
		return controllerutil.ReconcileErr(err)
	}
	return reconcile.Result{}, nil
}

func (r *reconciler) delete(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig) (reconcile.Result, error) {
	hasFinalizer, err := controllerutil.HasFinalizer(osc, FinalizerName)
	if err != nil {
		r.logger.Error(err, "Could not instantiate finalizer deletion")
		return reconcile.Result{}, err
	}

	if !hasFinalizer {
		r.logger.Info("Reconciling operating system config causes a no-op as there is no finalizer.", "osc", osc.Name)
		return reconcile.Result{}, nil
	}

	if err := r.actuator.Delete(ctx, osc); err != nil {
		r.logger.Error(err, "Error deleting operating system config", "osc", osc.Name)
		return controllerutil.ReconcileErr(err)
	}

	r.logger.Info("Operating system config deletion successful, removing finalizer.", "osc", osc.Name)
	if err := controllerutil.DeleteFinalizer(ctx, r.client, FinalizerName, osc); err != nil {
		r.logger.Error(err, "Error removing finalizer from operating system config", "osc", osc.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
