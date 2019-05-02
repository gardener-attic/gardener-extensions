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
	controllerconfig "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/controller/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ControllerName is the name of the lifecycle controller.
const ControllerName = "certificate-service-lifecycle-controller"

var (
	// ServiceConfig contains configuration for the certificate service.
	ServiceConfig controllerconfig.Config
)

// AddToManager adds a Certificate Service Lifecycle controller to the given Controller Manager.
func AddToManager(mgr manager.Manager) error {
	reconciler := NewReconciler(NewActuator(), ControllerName, ServiceConfig.Configuration)
	return mgr.Add(reconciler)
}
