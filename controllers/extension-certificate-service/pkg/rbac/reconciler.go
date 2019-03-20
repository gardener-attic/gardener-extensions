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

package rbac

import (
	"context"
	"fmt"
	"time"

	"github.com/gardener/gardener-extensions/pkg/controller/extension"

	"github.com/gardener/gardener-extensions/pkg/util"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	controllerutil "github.com/gardener/gardener-extensions/pkg/controller"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconciler reconciles Extension resources of Gardener's
// `extensions.gardener.cloud` API group.
type reconciler struct {
	actuator      Actuator
	logger        logr.Logger
	finalizerName string

	ctx               context.Context
	client            client.Client
	reconcileInterval time.Duration
}

var _ reconcile.Reconciler = (*reconciler)(nil)

// NewReconciler creates a new reconcile.Reconciler that reconciles
// Extension resources of Gardener's `extensions.gardener.cloud` API group.
func NewReconciler(actuator Actuator, name string, reconcileInterval time.Duration) reconcile.Reconciler {
	logger := log.Log.WithName(name)
	finalizer := fmt.Sprintf("%s/%s", extension.FinalizerPrefix, name)
	return &reconciler{
		logger:            logger,
		finalizerName:     finalizer,
		actuator:          actuator,
		reconcileInterval: reconcileInterval,
	}
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
	secret := &corev1.Secret{}
	if err := r.client.Get(r.ctx, request.NamespacedName, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		r.logger.Error(err, "Could not fetch Secret")
		return reconcile.Result{}, err
	}

	if _, ok := secret.Data["kubeconfig"]; !ok {
		return reconcile.Result{}, fmt.Errorf("no kubeconfig found in secret %s/%s", secret.Namespace, secret.Name)
	}

	config, err := clientcmd.RESTConfigFromKubeConfig(secret.Data["kubeconfig"])
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error creating kubeconfig: %v", err)
	}

	client, err := client.New(config, client.Options{})
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error creating kubeconfig: %v", err)
	}

	// Inject target cluster client into actuator.
	if s, ok := r.actuator.(inject.Config); ok {
		s.InjectConfig(config)
	}

	// Inject target cluster client into actuator.
	if s, ok := r.actuator.(inject.Client); ok {
		s.InjectClient(client)
	}

	if secret.DeletionTimestamp != nil {
		return r.delete(r.ctx, secret)
	}
	return r.reconcile(r.ctx, secret)
}

func (r *reconciler) reconcile(ctx context.Context, secret *corev1.Secret) (reconcile.Result, error) {
	if err := controllerutil.EnsureFinalizer(ctx, r.client, r.finalizerName, secret); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.actuator.Create(ctx); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true, RequeueAfter: r.reconcileInterval}, nil
}

func (r *reconciler) delete(ctx context.Context, secret *corev1.Secret) (reconcile.Result, error) {
	hasFinalizer, err := controllerutil.HasFinalizer(secret, r.finalizerName)
	if err != nil {
		r.logger.Error(err, "Could not instantiate finalizer deletion")
		return reconcile.Result{}, err
	}

	if !hasFinalizer {
		r.logger.Info("Reconciling Extension resource causes a no-op as there is no finalizer.", "secret", secret.Name, "namespace", secret.Namespace)
		return reconcile.Result{}, nil
	}

	if err := r.actuator.Delete(ctx); err != nil {
		return reconcile.Result{}, err
	}

	if err := controllerutil.DeleteFinalizer(ctx, r.client, r.finalizerName, secret); err != nil {
		r.logger.Error(err, "Error removing finalizer from Secret", "secret", secret.Name, "namespace", secret.Namespace)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
