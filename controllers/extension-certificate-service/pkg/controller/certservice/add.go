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
	"github.com/gardener/gardener-extensions/pkg/controller/extension"

	controllerconfig "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/controller/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// Type is the type of Extension resource.
	Type = "certificate-service"
	// ControllerName is the name of the Certificate Service controller.
	ControllerName = "certificate-service-controller"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the certificate service controller to the manager.
type AddOptions struct {
	// ControllerOptions contains options for the controller.
	ControllerOptions controller.Options
	// ServiceConfig contains configuration for the certificate service.
	ServiceConfig controllerconfig.Config
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
}

// AddToManager adds a controller with the default Options to the given Controller Manager.
func AddToManager(mgr manager.Manager) error {
	return AddToManagerWithOptions(mgr, DefaultAddOptions.ControllerOptions, DefaultAddOptions.ServiceConfig)
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
func AddToManagerWithOptions(mgr manager.Manager, opts controller.Options, config controllerconfig.Config) error {
	return extension.Add(mgr, extension.AddArgs{
		Actuator:          NewActuator(config.Configuration),
		ControllerOptions: opts,
		Name:              ControllerName,
		Resync:            config.Spec.ServiceSync.Duration,
		Predicates:        extension.DefaultPredicates(Type, DefaultAddOptions.IgnoreOperationAnnotation),
	})
}
