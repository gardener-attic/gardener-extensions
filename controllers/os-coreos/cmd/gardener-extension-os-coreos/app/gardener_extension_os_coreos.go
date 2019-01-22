/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This file was copied and modified from the kubernetes-csi/drivers project
https://github.com/kubernetes-sigs/controller-runtime/blob/v0.1.9/example/main.go

Modifications copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.
*/

package app

import (
	"flag"
	"fmt"
	"os"

	"github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig"

	"github.com/gardener/gardener-extensions/controllers/os-coreos/pkg/coreos"
	"github.com/gardener/gardener-extensions/controllers/os-coreos/pkg/version"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	extensionsscheme "github.com/gardener/gardener/pkg/client/extensions/clientset/versioned/scheme"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	corescheme "k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const controllerName = "gardener-extension-os-coreos"

var log = logf.Log.WithName(controllerName)

// NewCommandStartExtensionCoreOS starts the controller-manager for the Config Container Linux config extension
// for Gardener. It watches the `OperatingSystemConfig` resource in the `extensions.gardener.cloud` API group
// and reconciles it in case `.spec.type=coreos`.
func NewCommandStartExtensionCoreOS() error {
	var (
		ExtensionsScheme        = runtime.NewScheme()
		extensionsSchemeBuilder = runtime.NewSchemeBuilder(
			corescheme.AddToScheme,
			extensionsscheme.AddToScheme,
		)
	)
	utilruntime.Must(extensionsSchemeBuilder.AddToScheme(ExtensionsScheme))

	logf.SetLogger(logf.ZapLogger(false))
	entryLog := log.WithName("entrypoint")

	var concurrentSyncs int
	flag.IntVar(&concurrentSyncs, "concurrent-syncs", 5, "The number of syncing operations that will be done concurrently. Larger number = faster endpoint updating, but more CPU (and network) load.")
	flag.Parse()

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Scheme:                  ExtensionsScheme,
		LeaderElection:          true,
		LeaderElectionID:        fmt.Sprintf("%s-leader-election", controllerName),
		LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
	})
	if err != nil {
		entryLog.Error(err, "unable to set up overall controller manager")
		return err
	}

	c, err := controller.New(controllerName, mgr, controller.Options{
		MaxConcurrentReconciles: concurrentSyncs,
		Reconciler:              operatingsystemconfig.NewReconciler(log.WithName(controllerName), coreos.NewActuator(log.WithName(fmt.Sprintf("%s-actuator", controllerName)))),
	})
	if err != nil {
		entryLog.Error(err, "unable to set up individual controller")
		return err
	}

	if err := c.Watch(&source.Kind{Type: &extensionsv1alpha1.OperatingSystemConfig{}}, &handler.EnqueueRequestForObject{}, operatingsystemconfig.GenerationChangedPredicate(), coreos.Predicate()); err != nil {
		entryLog.Error(err, "unable to watch OperatingSystemConfig resources")
		return err
	}

	if err := c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: operatingsystemconfig.SecretToOSCMapper(mgr.GetClient())}); err != nil {
		entryLog.Error(err, "unable to watch Secret resources")
		return err
	}

	entryLog.Info("Starting controller manager", "version", version.Version)
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		entryLog.Error(err, "unable to run controller-manager")
		return err
	}

	return nil
}
