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

	controllererror "github.com/gardener/gardener-extensions/pkg/controller/error"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

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

// ContextFromStopChannel creates a new context from a given stop channel.
func ContextFromStopChannel(stopCh <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		<-stopCh
	}()

	return ctx
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
