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
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/certservice"
	controllerutil "github.com/gardener/gardener-extensions/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"sigs.k8s.io/controller-runtime/pkg/source"
)

func init() {
	addToManagerBuilder.Register(Add)
}

var (
	// Options are the default controller.Options for Add.
	Options = controller.Options{}

	// ReconcileInterval specifies the interval in which a request is requeued.
	ReconcileInterval = 5 * time.Minute
)

// Add adds a controller with the default Options to the given Controller Manager.
func Add(mgr manager.Manager) error {
	Options.Reconciler = NewReconciler(NewActuator(), ControllerName, ReconcileInterval)
	ctrl, err := controller.New(ControllerName, mgr, Options)
	if err != nil {
		return err
	}

	predicate := controllerutil.NamePredicate(certservice.RBACManagerSecretName)

	// Add standard watch.
	if err := ctrl.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{}, predicate); err != nil {
		return err
	}

	return nil
}
