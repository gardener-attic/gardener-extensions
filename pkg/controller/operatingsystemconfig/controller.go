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

	extensionspredicate "github.com/gardener/gardener-extensions/pkg/predicate"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// FinalizerName is the name of the finalizer written by this controller.
	FinalizerName = "extensions.gardener.cloud/operatingsystemconfigs"

	// ControllerName is the name of the operating system configuration controller.
	ControllerName = "operatingsystemconfig_controller"

	name = "operatingsystemconfig-controller"

	specFilesSecretFieldKey = "secretRef.name"
)

// AddArgs are arguments for adding an operatingsystemconfig controller to a manager.
type AddArgs struct {
	// Actuator is an operatingsystemconfig actuator.
	Actuator Actuator
	// ControllerOptions are the controller options used for creating a controller.
	// The options.Reconciler is always overridden with a reconciler created from the
	// given actuator.
	ControllerOptions controller.Options
	// Predicates are the predicates to use.
	// If unset, GenerationChanged will be used.
	Predicates []predicate.Predicate
}

// Add adds an operatingsystemconfig controller to the given manager using the given AddArgs.
func Add(mgr manager.Manager, args AddArgs) error {
	args.ControllerOptions.Reconciler = NewReconciler(args.Actuator)
	return add(mgr, args.ControllerOptions, args.Predicates)
}

// DefaultPredicates returns the default predicates for an operatingsystemconfig reconciler.
func DefaultPredicates(typeName string, ignoreOperationAnnotation bool) []predicate.Predicate {
	if ignoreOperationAnnotation {
		return []predicate.Predicate{
			extensionspredicate.HasType(typeName),
			extensionspredicate.GenerationChanged(),
		}
	}

	return []predicate.Predicate{
		extensionspredicate.HasType(typeName),
		extensionspredicate.Or(
			extensionspredicate.HasOperationAnnotation(),
			extensionspredicate.LastOperationNotSuccessful(),
			extensionspredicate.IsDeleting(),
		),
	}
}

func add(mgr manager.Manager, options controller.Options, predicates []predicate.Predicate) error {
	logger := log.Log.WithName(name)
	ctrl, err := controller.New(ControllerName, mgr, options)
	if err != nil {
		return err
	}

	if err := ctrl.Watch(&source.Kind{Type: &extensionsv1alpha1.OperatingSystemConfig{}}, &handler.EnqueueRequestForObject{}, predicates...); err != nil {
		return err
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &corev1.Secret{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {

				cl := mgr.GetClient()
				oscLists := &extensionsv1alpha1.OperatingSystemConfigList{}
				if err := cl.List(
					context.Background(),
					oscLists,
					client.InNamespace(a.Meta.GetNamespace()),
					client.MatchingField(specFilesSecretFieldKey, a.Meta.GetName())); err != nil {
					logger.Error(err, "cannot list OperatingSystemConfig with fieldselector", "name", a.Meta.GetName(), "namespace", a.Meta.GetNamespace())
					return []reconcile.Request{}
				}

				result := make([]reconcile.Request, 0, len(oscLists.Items))
				for _, osc := range oscLists.Items {
					result = append(result, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      osc.GetName(),
							Namespace: osc.GetNamespace(),
						},
					})
				}
				return result
			}),
		}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(&extensionsv1alpha1.OperatingSystemConfig{}, specFilesSecretFieldKey, func(rawObj runtime.Object) []string {
		osc, ok := rawObj.(*extensionsv1alpha1.OperatingSystemConfig)
		if !ok || osc == nil {
			return nil
		}
		secrets := sets.String{}
		for _, f := range osc.Spec.Files {
			if sr := f.Content.SecretRef; sr != nil {
				secrets.Insert(sr.Name)
			}
		}
		if cc := osc.Status.CloudConfig; cc != nil {
			secrets.Insert(cc.SecretRef.Name)
		}
		return secrets.List()
	}); err != nil {
		return err
	}

	return nil
}
