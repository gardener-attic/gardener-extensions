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
	"time"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config"
	"github.com/gardener/gardener-extensions/pkg/util"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// Reconciler reconciles CertificateServiceConfig related to the certificate service.
type Reconciler struct {
	logger   logr.Logger
	actuator Actuator

	ctx    context.Context
	client client.Client

	certServiceConfig config.Configuration
}

// NewReconciler creates a new Reconciler.
func NewReconciler(actuator Actuator, name string, config config.Configuration) *Reconciler {
	logger := log.Log.WithName(name)
	return &Reconciler{
		logger:            logger,
		actuator:          actuator,
		certServiceConfig: config,
	}
}

// Start implements manager.Runnable interface.
func (r *Reconciler) Start(stop <-chan struct{}) error {
	retryBackoff := wait.Backoff{
		Duration: 10 * time.Second,
		Factor:   2.0,
		Steps:    5,
	}

	wait.Until(func() {
		_ = wait.ExponentialBackoff(retryBackoff, func() (done bool, err error) {
			if err := r.Reconcile(r.certServiceConfig); err != nil {
				return false, nil
			}
			return true, nil
		})
	}, r.certServiceConfig.Spec.LifecycleSync.Duration, stop)

	return nil
}

// InjectFunc enables dependency injection into the actuator.
func (r *Reconciler) InjectFunc(f inject.Func) error {
	return f(r.actuator)
}

// InjectClient injects the controller runtime client into the reconciler.
func (r *Reconciler) InjectClient(client client.Client) error {
	r.client = client
	return nil
}

// InjectStopChannel is an implementation for getting the respective stop channel managed by the controller-runtime.
func (r *Reconciler) InjectStopChannel(stopCh <-chan struct{}) error {
	r.ctx = util.ContextFromStopChannel(stopCh)
	return nil
}

// Reconcile is the reconciler function that gets executed in case there are new events for the `Configuration`
// which contains the configuration for the certificate service.
func (r *Reconciler) Reconcile(config config.Configuration) error {
	return r.reconcile(r.ctx, config)
}

func (r *Reconciler) reconcile(ctx context.Context, config config.Configuration) error {
	r.logger.Info("Reconciling certificate service configuration triggers idempotent create or update")
	if err := r.actuator.Reconcile(ctx, config); err != nil {
		r.logger.Error(err, "Reconciliation failed for certificate service")
		return err
	}
	return nil
}
