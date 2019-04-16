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

	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// newMutator creates a new controlplane mutator.
func newMutator(logger logr.Logger) *mutator {
	return &mutator{
		logger: logger.WithName("mutator"),
	}
}

type mutator struct {
	logger logr.Logger
}

// Mutate validates and if needed mutates the given object.
func (v *mutator) Mutate(ctx context.Context, obj runtime.Object) error {
	switch x := obj.(type) {
	case *appsv1.Deployment:
		switch x.Name {
		case common.KubeAPIServerDeploymentName:
			return mutateKubeAPIServerDeployment(x)
		case common.KubeControllerManagerDeploymentName:
			return mutateKubeControllerManagerDeployment(x)
		}
	}
	return nil
}

func mutateKubeAPIServerDeployment(obj *appsv1.Deployment) error {
	// TODO
	return nil
}

func mutateKubeControllerManagerDeployment(obj *appsv1.Deployment) error {
	// TODO
	return nil
}
