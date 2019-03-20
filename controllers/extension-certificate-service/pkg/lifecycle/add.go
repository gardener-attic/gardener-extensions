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
	configv1alpha1 "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config/v1alpha1"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"time"

	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func init() {
	addToManagerBuilder.Register(Add)
}

var (
	// Options are the default controller.Options for Add.
	Options = controller.Options{
		// Since one config is reconciled, only one worker is required.
		MaxConcurrentReconciles: 1,
	}

	// ResourceNamespace is the namespace used for reading configuration and creating resources.
	ResourceNamespace = "kube-system"

	// ResourceName is the name of the CertificateServiceConfiguration.
	ResourceName = "certificate-service"

	// Sync determins the time the configuration of the certificate-service is reconciled.
	Sync = 1 * time.Hour
)

// Add adds a Certificate Service Lifecycle controller to the given Controller Manager.
func Add(mgr manager.Manager) error {
	Options.Reconciler = NewReconciler(NewActuator(ResourceNamespace), ControllerName, Sync)
	ctrl, err := controller.New(ControllerName, mgr, Options)
	if err != nil {
		return err
	}

	predicates := []predicate.Predicate{
		extensionscontroller.NameAndNamespacePredicate(ResourceName, ResourceNamespace),
		extensionscontroller.OrPredicate(
			extensionscontroller.GenerationChangedPredicate(),
			// Predicate used to catch update event which is triggered upon deleting a resource with finalizers.
			extensionscontroller.HasDeletionTimestampPredicate(),
		),
	}

	if err := ctrl.Watch(&source.Kind{Type: &configv1alpha1.CertificateServiceConfiguration{}}, &handler.EnqueueRequestForObject{}, predicates...); err != nil {
		return err
	}

	return nil
}
