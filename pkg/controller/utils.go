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

package controller

import (
	"context"
	"github.com/gardener/gardener-extensions/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	controllererror "github.com/gardener/gardener-extensions/pkg/controller/error"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

import (
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		scheme.AddToScheme,
		extensionsv1alpha1.AddToScheme,
	)

	// AddToScheme adds the Kubernetes and extension scheme to the given scheme.
	AddToScheme = localSchemeBuilder.AddToScheme

	// ExtensionsScheme is the default scheme for extensions, consisting of all Kubernetes built-in
	// schemes (client-go/kubernetes/scheme) and the extensions/v1alpha1 scheme.
	ExtensionsScheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(AddToScheme(ExtensionsScheme))
}

// ReconcileErr returns a reconcile.Result or an error, depending on whether the error is a
// RequeueAfterError or not.
func ReconcileErr(err error) (reconcile.Result, error) {
	if requeueAfter, ok := err.(*controllererror.RequeueAfterError); ok {
		return reconcile.Result{Requeue: true, RequeueAfter: requeueAfter.RequeueAfter}, nil
	}
	return reconcile.Result{}, err
}

// LastOperation creates a new LastOperation from the given parameters.
func LastOperation(t extensionsv1alpha1.LastOperationType, state extensionsv1alpha1.LastOperationState, progress int, description string) *extensionsv1alpha1.LastOperation {
	return &extensionsv1alpha1.LastOperation{
		LastUpdateTime: metav1.Now(),
		Type:           t,
		State:          state,
		Description:    description,
		Progress:       progress,
	}
}

// LastError creates a new LastError from the given parameters.
func LastError(description string, codes ...extensionsv1alpha1.ErrorCode) *extensionsv1alpha1.LastError {
	return &extensionsv1alpha1.LastError{
		Description: description,
		Codes:       codes,
	}
}

// ReconcileSucceeded returns a LastOperation with state succeeded at 100 percent and a nil LastError.
func ReconcileSucceeded(t extensionsv1alpha1.LastOperationType, description string) (*extensionsv1alpha1.LastOperation, *extensionsv1alpha1.LastError) {
	return LastOperation(t, extensionsv1alpha1.LastOperationStateSucceeded, 100, description), nil
}

// ReconcileError returns a LastOperation with state error and a LastError with the given description and codes.
func ReconcileError(t extensionsv1alpha1.LastOperationType, description string, progress int, codes ...extensionsv1alpha1.ErrorCode) (*extensionsv1alpha1.LastOperation, *extensionsv1alpha1.LastError) {
	return LastOperation(t, extensionsv1alpha1.LastOperationStateError, progress, description), LastError(description, codes...)
}

// AddToManagerBuilder aggregates various AddToManager functions.
type AddToManagerBuilder []func(manager.Manager) error

// NewAddToManagerBuilder creates a new AddToManagerBuilder and registers the given functions.
func NewAddToManagerBuilder(funcs ...func(manager.Manager) error) AddToManagerBuilder {
	var builder AddToManagerBuilder
	builder.Register(funcs...)
	return builder
}

// Register registers the given functions in this builder.
func (a *AddToManagerBuilder) Register(funcs ...func(manager.Manager) error) {
	*a = append(*a, funcs...)
}

// AddToManager traverses over all AddToManager-functions of this builder, sequentially applying
// them. It exits on the first error and returns it.
func (a *AddToManagerBuilder) AddToManager(m manager.Manager) error {
	for _, f := range *a {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}

func finalizersAndAccessorOf(obj runtime.Object) (sets.String, metav1.Object, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, nil, err
	}

	return sets.NewString(accessor.GetFinalizers()...), accessor, nil
}

// HasFinalizer checks if the given object has a finalizer with the given name.
func HasFinalizer(obj runtime.Object, finalizerName string) (bool, error) {
	finalizers, _, err := finalizersAndAccessorOf(obj)
	if err != nil {
		return false, err
	}

	return finalizers.Has(finalizerName), nil
}

// EnsureFinalizer ensures that a finalizer of the given name is set on the given object.
// If the finalizer is not set, it adds it to the list of finalizers and updates the remote object.
func EnsureFinalizer(ctx context.Context, client client.Client, finalizerName string, obj runtime.Object) error {
	finalizers, accessor, err := finalizersAndAccessorOf(obj)
	if err != nil {
		return err
	}

	if finalizers.Has(finalizerName) {
		return nil
	}

	finalizers.Insert(finalizerName)
	accessor.SetFinalizers(finalizers.UnsortedList())

	return client.Update(ctx, obj)
}

// DeleteFinalizer ensures that the given finalizer is not present anymore in the given object.
// If it is set, it removes it and issues an update.
func DeleteFinalizer(ctx context.Context, client client.Client, finalizerName string, obj runtime.Object) error {
	finalizers, accessor, err := finalizersAndAccessorOf(obj)
	if err != nil {
		return err
	}

	if !finalizers.Has(finalizerName) {
		return nil
	}

	finalizers.Delete(finalizerName)
	accessor.SetFinalizers(finalizers.UnsortedList())

	return client.Update(ctx, obj)
}

// CreateOrUpdate creates or updates the object. Optionally, it executes a transformation function before the
// request is made.
func CreateOrUpdate(ctx context.Context, c client.Client, obj runtime.Object, transform func() error) error {
	key, err := client.ObjectKeyFromObject(obj)
	if err != nil {
		return err
	}

	if err := c.Get(ctx, key, obj); err != nil {
		if apierrors.IsNotFound(err) {
			if transform != nil && transform() != nil {
				return err
			}
			return c.Create(ctx, obj)
		}
		return err
	}

	if transform != nil && transform() != nil {
		return err
	}
	return c.Update(ctx, obj)
}

// SetupSignalHandlerContext sets up a context from signals.SetupSignalHandler stop channel.
func SetupSignalHandlerContext() context.Context {
	return util.ContextFromStopChannel(signals.SetupSignalHandler())
}
