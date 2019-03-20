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

package lifecycle

import (
	"context"
	"fmt"
	"time"

	"github.com/gardener/gardener-extensions/pkg/controller/extension"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	configv1alpha1 "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config/v1alpha1"
)

// reconciler reconciles CertificateServiceConfig related to the certificate service.
type reconciler struct {
	logger        logr.Logger
	actuator      Actuator
	finalizerName string
	sync          time.Duration

	ctx    context.Context
	client client.Client
}

var _ reconcile.Reconciler = (*reconciler)(nil)

// NewReconciler creates a new reconcile.Reconciler that reconciles CertificateServiceConfig.
func NewReconciler(actuator Actuator, name string, sync time.Duration) reconcile.Reconciler {
	logger := log.Log.WithName(name)
	finalizer := fmt.Sprintf("%s/%s", extension.FinalizerPrefix, name)
	return &reconciler{
		logger:        logger,
		actuator:      actuator,
		finalizerName: finalizer,
		sync:          sync,
	}
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

// Reconcile is the reconciler function that gets executed in case there are new events for the `CertificateServiceConfiguration`
// which contains the configuration for the certificate service.
func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	config := &configv1alpha1.CertificateServiceConfiguration{}
	if err := r.client.Get(r.ctx, request.NamespacedName, config); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		r.logger.Error(err, "Could not fetch CertificateServiceConfig", "CertificateServiceConfiguration", request.NamespacedName.Name, "namespace", request.NamespacedName.Namespace)
		return reconcile.Result{Requeue: true}, err
	}

	if config.DeletionTimestamp != nil {
		return r.delete(r.ctx, config)
	}
	return r.reconcile(r.ctx, config)
}

func (r *reconciler) reconcile(ctx context.Context, config *configv1alpha1.CertificateServiceConfiguration) (reconcile.Result, error) {
	if err := extensionscontroller.EnsureFinalizer(ctx, r.client, r.finalizerName, config); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Reconciling certificate service configuration triggers idempotent create or update", "CertificateServiceConfiguration", config.Name)
	if err := r.actuator.CreateOrUpdate(ctx, config); err != nil {
		r.logger.Error(err, "Reconcilation failed for certificate service", "CertificateServiceConfiguration", config.Name, "namespace", config.Namespace)
		return extensionscontroller.ReconcileErr(err)
	}
	return reconcile.Result{RequeueAfter: r.sync}, nil
}

func (r *reconciler) delete(ctx context.Context, config *configv1alpha1.CertificateServiceConfiguration) (reconcile.Result, error) {
	hasFinalizer, err := extensionscontroller.HasFinalizer(config, r.finalizerName)
	if err != nil {
		r.logger.Error(err, "Could not instantiate finalizer deletion", "CertificateServiceConfiguration", config.Name, "namespace", config.Namespace)
		return reconcile.Result{}, err
	}

	if !hasFinalizer {
		r.logger.Info("Reconciling certificate service configuration causes a no-op as there is no finalizer", "CertificateServiceConfiguration", config.Name, "namespace", config.Namespace)
		return reconcile.Result{}, nil
	}

	if err := r.actuator.Delete(ctx, config); err != nil {
		r.logger.Error(err, "Error deleting certificate service configuration", "CertificateServiceConfiguration", config.Name, "namespace", config.Namespace)
		return extensionscontroller.ReconcileErr(err)
	}

	r.logger.Info("Certificate service configuration deletion successful, removing finalizer.", "CertificateServiceConfiguration", config.Name, "namespace", config.Namespace)
	if err := extensionscontroller.DeleteFinalizer(ctx, r.client, r.finalizerName, config); err != nil {
		r.logger.Error(err, "Error removing finalizer from Certificate service configuration", "CertificateServiceConfiguration", config.Name, "namespace", config.Namespace)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
