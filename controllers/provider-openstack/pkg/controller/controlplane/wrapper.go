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

package controlplane

import (
	"context"
	"time"

	"github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane"

	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

type wrapper struct {
	controlplane.Actuator
}

func (w *wrapper) InjectFunc(f inject.Func) error {
	return f(w.Actuator)
}

func (w *wrapper) Delete(ctx context.Context, cp *v1alpha1.ControlPlane, cluster *controller.Cluster) error {
	// TODO: Dirty fix. Nothing to see here. This needs to be refactored!!!!
	// 		 In the future use gophercloud to check whether there are LoadBalancers which belong to any
	// 		 Shoot worker subnet work (have ports in the worker subnetwork).
	time.Sleep(2 * time.Minute)
	return w.Actuator.Delete(ctx, cp, cluster)
}
