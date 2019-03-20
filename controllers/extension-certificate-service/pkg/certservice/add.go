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

package certservice

import (
	"context"

	"github.com/gardener/gardener-extensions/pkg/controller/extension"

	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1alpha1 "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config/v1alpha1"
	controllerutil "github.com/gardener/gardener-extensions/pkg/controller"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	corev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Type is the type of OperatingSystemConfigs the coreos actuator / predicate are built for.
const Type = "certificate-service"

func init() {
	addToManagerBuilder.Register(Add)
}

var (
	// Options are the default controller.Options for Add.
	Options = controller.Options{}

	// ResourceNamespace is the namespace used for reading configuration and creating resources.
	ResourceNamespace = "kube-system"

	// ResourceName is the name of the CertServiceConfig.
	ResourceName = "certificate-service"

	// watchBuilder determines which resources should be watched and the reconciler to be used.
	watchBuilder controllerutil.WatcherBuilder
)

// Add adds a controller with the default Options to the given Controller Manager.
func Add(mgr manager.Manager) error {
	return AddWithOptions(mgr, Options)
}

// AddWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
func AddWithOptions(mgr manager.Manager, opts controller.Options) error {
	var (
		cl                 = mgr.GetClient()
		mapper             = typeMapperWithinNamespace(cl)
		allNamespaceMapper = typeMapperForAllNamespaces(cl)
	)

	// Register Watch on Cluster resources.
	watchBuilder.Register(func(ctrl controller.Controller) error {
		if err := ctrl.Watch(
			&source.Kind{Type: &extensionsv1alpha1.Cluster{}},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: mapper},
		); err != nil {
			return err
		}
		return nil
	})

	// Register Watch on CA Secret resource.
	watchBuilder.Register(func(ctrl controller.Controller) error {
		if err := ctrl.Watch(
			&source.Kind{Type: &corev1.Secret{}},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: mapper},
			controllerutil.NamePredicate(corev1alpha1.SecretNameCACluster),
		); err != nil {
			return err
		}
		return nil
	})

	// Register Watch on CertificateServiceConfiguration.
	watchBuilder.Register(func(ctrl controller.Controller) error {
		if err := ctrl.Watch(
			&source.Kind{Type: &configv1alpha1.CertificateServiceConfiguration{}},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: allNamespaceMapper},
			extensionscontroller.NameAndNamespacePredicate(ResourceName, ResourceNamespace),
			extensionscontroller.GenerationChangedPredicate(),
		); err != nil {
			return err
		}
		return nil
	})

	return extension.Add(mgr, extension.AddArgs{
		Actuator:          NewActuator(),
		Name:              ControllerName,
		Type:              Type,
		ControllerOptions: opts,
		WatchBuilder:      watchBuilder,
	})
}

// typeMapperWithinNamespace returns a `ToRequestsFunc` that maps the incoming object
// to the certificate service extension object in the same namespace.
func typeMapperWithinNamespace(cl client.Client) handler.ToRequestsFunc {
	return func(object handler.MapObject) []reconcile.Request {
		geList := extensionsv1alpha1.ExtensionList{}
		if err := cl.List(context.TODO(), client.InNamespace(object.Meta.GetNamespace()), &geList); err != nil {
			return nil
		}

		var resourceName string
		for _, ge := range geList.Items {
			if controllerutil.EvalGenericPredicate(&ge, controllerutil.TypePredicate(Type)) {
				resourceName = ge.GetName()
				break
			}
		}

		return []reconcile.Request{
			reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      resourceName,
					Namespace: object.Meta.GetNamespace(),
				},
			}}
	}
}

// typeMapperForAllNamespaces returns a `ToRequestsFunc` that maps the incoming object
// to the certificate service extension objects across all namespaces.
func typeMapperForAllNamespaces(cl client.Client) handler.ToRequestsFunc {
	return func(object handler.MapObject) []reconcile.Request {
		geList := extensionsv1alpha1.ExtensionList{}
		if err := cl.List(context.TODO(), &client.ListOptions{}, &geList); err != nil {
			return nil
		}

		var requests []reconcile.Request
		for _, ge := range geList.Items {
			if controllerutil.EvalGenericPredicate(&ge, controllerutil.TypePredicate(Type)) {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      ge.GetName(),
						Namespace: ge.GetNamespace(),
					},
				})
			}
		}

		return requests
	}
}
