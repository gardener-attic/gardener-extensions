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

package extension

import (
	"context"
	"fmt"
	extensionspredicate "github.com/gardener/gardener-extensions/pkg/predicate"
	"time"

	"k8s.io/client-go/util/retry"

	"github.com/gardener/gardener-extensions/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	extensionshandler "github.com/gardener/gardener-extensions/pkg/handler"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	gardencorev1alpha1helper "github.com/gardener/gardener/pkg/apis/core/v1alpha1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// FinalizerPrefix is the prefix name of the finalizer written by this controller.
	FinalizerPrefix = "extensions.gardener.cloud"
)

// AddArgs are arguments for adding an Extension resources controller to a manager.
type AddArgs struct {
	// Actuator is an Extension resource actuator.
	Actuator Actuator
	// Name is the name of the controller.
	Name string
	// ControllerOptions are the controller options used for creating a controller.
	// The options.Reconciler is always overridden with a reconciler created from the
	// given actuator.
	ControllerOptions controller.Options
	// Predicates are the predicates to use.
	Predicates []predicate.Predicate
	// Resync determines the requeue interval.
	Resync time.Duration
}

// Add adds an Extension controller to the given manager using the given AddArgs.
func Add(mgr manager.Manager, args AddArgs) error {
	args.ControllerOptions.Reconciler = NewReconciler(args)
	return add(mgr, args)
}

// DefaultPredicates returns the default predicates for an extension reconciler.
func DefaultPredicates(extensionType string, ignoreOperationAnnotation bool) []predicate.Predicate {
	if ignoreOperationAnnotation {
		return []predicate.Predicate{
			extensionspredicate.HasType(extensionType),
			extensionspredicate.GenerationChanged(),
		}
	}
	return []predicate.Predicate{
		extensionspredicate.HasType(extensionType),
		extensionspredicate.Or(
			extensionspredicate.HasOperationAnnotation(),
			extensionspredicate.LastOperationNotSuccessful(),
			extensionspredicate.IsDeleting(),
		),
		extensionspredicate.ShootNotFailed(),
	}
}

func add(mgr manager.Manager, args AddArgs) error {
	ctrl, err := controller.New(args.Name, mgr, args.ControllerOptions)
	if err != nil {
		return err
	}

	if err := ctrl.Watch(&source.Kind{Type: &extensionsv1alpha1.Extension{}}, &handler.EnqueueRequestForObject{}, args.Predicates...); err != nil {
		return err
	}

	return ctrl.Watch(&source.Kind{Type: &extensionsv1alpha1.Cluster{}}, &extensionshandler.EnqueueRequestsFromMapFunc{
		ToRequests: extensionshandler.SimpleMapper(ClusterToExtensionMapper(mgr.GetClient(), args.Predicates...), extensionshandler.UpdateWithNew),
	})
}

// reconciler reconciles Extension resources of Gardener's
// `extensions.gardener.cloud` API group.
type reconciler struct {
	logger        logr.Logger
	actuator      Actuator
	finalizerName string

	ctx    context.Context
	client client.Client

	resync time.Duration
}

var _ reconcile.Reconciler = (*reconciler)(nil)

// NewReconciler creates a new reconcile.Reconciler that reconciles
// Extension resources of Gardener's `extensions.gardener.cloud` API group.
func NewReconciler(args AddArgs) reconcile.Reconciler {
	logger := log.Log.WithName(args.Name)
	finalizer := fmt.Sprintf("%s/%s", FinalizerPrefix, args.Name)
	return extensionscontroller.OperationAnnotationWrapper(
		&extensionsv1alpha1.Extension{},
		&reconciler{
			logger:        logger,
			actuator:      args.Actuator,
			finalizerName: finalizer,
			resync:        args.Resync,
		})
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

// Reconcile is the reconciler function that gets executed in case there are new events for `Extension` resources.
func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ex := &extensionsv1alpha1.Extension{}
	if err := r.client.Get(r.ctx, request.NamespacedName, ex); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		r.logger.Error(err, "Could not fetch Extension resource")
		return reconcile.Result{}, err
	}

	var (
		result reconcile.Result
		err    error
	)

	if ex.DeletionTimestamp != nil {
		return r.delete(r.ctx, ex)
	}

	result, err = r.reconcile(r.ctx, ex)
	if err != nil {
		return result, err
	}
	return reconcile.Result{Requeue: r.resync != 0, RequeueAfter: r.resync}, nil
}

func (r *reconciler) reconcile(ctx context.Context, ex *extensionsv1alpha1.Extension) (reconcile.Result, error) {
	if err := extensionscontroller.EnsureFinalizer(ctx, r.client, r.finalizerName, ex); err != nil {
		return reconcile.Result{}, err
	}

	msg := "Reconciling Extension resource"
	r.logger.Info("Reconciling Extension resource", "extension", ex.Name, "namespace", ex.Namespace)
	operationType := gardencorev1alpha1helper.ComputeOperationType(ex.ObjectMeta, ex.Status.LastOperation)
	if err := r.updateStatusProcessing(ctx, ex, operationType, msg); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.actuator.Reconcile(ctx, ex); err != nil {
		msg := "Unable to reconcile Extension resource"
		_ = r.updateStatusError(ctx, extensionscontroller.ReconcileErrCauseOrErr(err), ex, operationType, msg)
		r.logger.Error(err, msg, "extension", ex.Name, "namespace", ex.Namespace)
		return extensionscontroller.ReconcileErr(err)
	}

	msg = "Successfully reconciled Extension resource"
	r.logger.Info(msg, "extension", ex.Name, "namespace", ex.Namespace)
	if err := r.updateStatusSuccess(ctx, ex, operationType, msg); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *reconciler) delete(ctx context.Context, ex *extensionsv1alpha1.Extension) (reconcile.Result, error) {
	hasFinalizer, err := extensionscontroller.HasFinalizer(ex, r.finalizerName)
	if err != nil {
		r.logger.Error(err, "Could not instantiate finalizer deletion")
		return reconcile.Result{}, err
	}

	if !hasFinalizer {
		r.logger.Info("Reconciling Extension resource causes a no-op as there is no finalizer.", "extension", ex.Name, "namespace", ex.Namespace)
		return reconcile.Result{}, nil
	}

	operationType := gardencorev1alpha1helper.ComputeOperationType(ex.ObjectMeta, ex.Status.LastOperation)
	if err := r.updateStatusProcessing(ctx, ex, operationType, "Deleting Extension resource."); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.actuator.Delete(ctx, ex); err != nil {
		msg := "Error deleting Extension resource"
		_ = r.updateStatusError(ctx, extensionscontroller.ReconcileErrCauseOrErr(err), ex, operationType, msg)
		r.logger.Error(err, msg, "extension", ex.Name, "namespace", ex.Namespace)
		return extensionscontroller.ReconcileErr(err)
	}

	msg := "Successfully deleted Extension resource"
	r.logger.Info(msg, "extension", ex.Name, "namespace", ex.Namespace)
	if err := r.updateStatusSuccess(ctx, ex, operationType, msg); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Extension resource deletion successful, removing finalizer.", "extension", ex.Name, "namespace", ex.Namespace)
	if err := extensionscontroller.DeleteFinalizer(ctx, r.client, r.finalizerName, ex); err != nil {
		r.logger.Error(err, "Error removing finalizer from Extension resource", "extension", ex.Name, "namespace", ex.Namespace)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *reconciler) updateStatusProcessing(ctx context.Context, ex *extensionsv1alpha1.Extension, lastOperationType gardencorev1alpha1.LastOperationType, description string) error {
	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, r.client, ex, func() error {
		ex.Status.LastOperation = extensionscontroller.LastOperation(lastOperationType, gardencorev1alpha1.LastOperationStateProcessing, 1, description)
		return nil
	})
}

func (r *reconciler) updateStatusError(ctx context.Context, err error, ex *extensionsv1alpha1.Extension, lastOperationType gardencorev1alpha1.LastOperationType, description string) error {
	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, r.client, ex, func() error {
		ex.Status.ObservedGeneration = ex.Generation
		ex.Status.LastOperation, ex.Status.LastError = extensionscontroller.ReconcileError(lastOperationType, gardencorev1alpha1helper.FormatLastErrDescription(fmt.Errorf("%s: %v", description, err)), 50, gardencorev1alpha1helper.ExtractErrorCodes(err)...)
		return nil
	})
}

func (r *reconciler) updateStatusSuccess(ctx context.Context, ex *extensionsv1alpha1.Extension, lastOperationType gardencorev1alpha1.LastOperationType, description string) error {
	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, r.client, ex, func() error {
		ex.Status.ObservedGeneration = ex.Generation
		ex.Status.LastOperation, ex.Status.LastError = extensionscontroller.ReconcileSucceeded(lastOperationType, description)
		return nil
	})
}
