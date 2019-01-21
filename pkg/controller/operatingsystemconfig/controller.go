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

	"github.com/gardener/gardener-extensions/pkg/controller"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// FinalizerName is the name of the finalizer written by this controller.
	FinalizerName = "extensions.gardener.cloud/operatingsystemconfigs"
)

// operatingSystemConfigReconciler reconciles OperatingSystemConfig resources of Gardener's
// `extensions.gardener.cloud` API group.
type operatingSystemConfigReconciler struct {
	logger   logr.Logger
	actuator Actuator

	ctx    context.Context
	client client.Client
}

var _ reconcile.Reconciler = &operatingSystemConfigReconciler{}

// NewReconciler creates a new reconcile.Reconciler that reconciles
// OperatingSystemConfig resources of Gardener's `extensions.gardener.cloud` API group.
func NewReconciler(logger logr.Logger, actuator Actuator) reconcile.Reconciler {
	return &operatingSystemConfigReconciler{logger: logger, actuator: actuator}
}

// InjectFunc enables dependency injection into the actuator.
func (r *operatingSystemConfigReconciler) InjectFunc(f inject.Func) error {
	return f(r.actuator)
}

// InjectClient injects the controller runtime client into the reconciler.
func (r *operatingSystemConfigReconciler) InjectClient(client client.Client) error {
	r.client = client
	return nil
}

// InjectStopChannel is an implementation for getting the respective stop channel managed by the controller-runtime.
func (r *operatingSystemConfigReconciler) InjectStopChannel(stopCh <-chan struct{}) error {
	r.ctx = controller.ContextFromStopChannel(stopCh)
	return nil
}

// Reconcile is the reconciler function that gets executed in case there are new events for the `OperatingSystemConfig`
// resources.
func (r *operatingSystemConfigReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
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

func (r *operatingSystemConfigReconciler) reconcile(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig) (reconcile.Result, error) {
	// Add finalizer to resource if not yet done.
	if finalizers := sets.NewString(osc.Finalizers...); !finalizers.Has(FinalizerName) {
		finalizers.Insert(FinalizerName)
		osc.Finalizers = finalizers.UnsortedList()
		if err := r.client.Update(ctx, osc); err != nil {
			return reconcile.Result{}, err
		}
	}

	exist, err := r.actuator.Exists(ctx, osc)
	if err != nil {
		return reconcile.Result{}, err
	}

	if exist {
		r.logger.Info("Reconciling operating system config triggers idempotent update.", "osc", osc.Name)
		if err := r.actuator.Update(ctx, osc); err != nil {
			return controller.ReconcileErr(err)
		}
		return reconcile.Result{}, nil
	}

	r.logger.Info("Reconciling operating system config triggers idempotent create.", "osc", osc.Name)
	if err := r.actuator.Create(ctx, osc); err != nil {
		r.logger.Error(err, "Unable to create operating system config", "osc", osc.Name)
		return controller.ReconcileErr(err)
	}
	return reconcile.Result{}, nil
}

func (r *operatingSystemConfigReconciler) delete(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig) (reconcile.Result, error) {
	finalizers := sets.NewString(osc.Finalizers...)
	if !finalizers.Has(FinalizerName) {
		r.logger.Info("Reconciling operating system config causes a no-op as there is no finalizer.", "osc", osc.Name)
		return reconcile.Result{}, nil
	}

	if err := r.actuator.Delete(ctx, osc); err != nil {
		r.logger.Error(err, "Error deleting operating system config", "osc", osc.Name)
		return controller.ReconcileErr(err)
	}

	r.logger.Info("Operating system config deletion successful, removing finalizer.", "osc", osc.Name)
	finalizers.Delete(FinalizerName)
	osc.Finalizers = finalizers.UnsortedList()
	if err := r.client.Update(ctx, osc); err != nil {
		r.logger.Error(err, "Error removing finalizer from operating system config", "osc", osc.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
