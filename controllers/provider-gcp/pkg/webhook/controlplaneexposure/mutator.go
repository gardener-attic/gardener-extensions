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

package controlplaneexposure

import (
	"context"

	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"

	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewMutator creates a new controlplaneexposure mutator.
func NewMutator(logger logr.Logger) controlplane.Mutator {
	return &mutator{
		logger: logger.WithName("mutator"),
	}
}

type mutator struct {
	client client.Client
	logger logr.Logger
}

// InjectClient injects the given client into the mutator.
func (m *mutator) InjectClient(client client.Client) error {
	m.client = client
	return nil
}

// Mutate validates and if needed mutates the given object.
func (m *mutator) Mutate(ctx context.Context, obj runtime.Object) error {
	switch x := obj.(type) {
	case *appsv1.Deployment:
		switch x.Name {
		case common.KubeAPIServerDeploymentName:
			// Get load balancer address of the kube-apiserver service
			address, err := controlplane.GetLoadBalancerIngress(ctx, m.client, x.Namespace, common.KubeAPIServerDeploymentName)
			if err != nil {
				return errors.Wrap(err, "could not get kube-apiserver service load balancer address")
			}

			return mutateKubeAPIServerDeployment(x, address)
		}
	}
	return nil
}

func mutateKubeAPIServerDeployment(dep *appsv1.Deployment, address string) error {
	if c := controlplane.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-apiserver"); c != nil {
		c.Command = controlplane.EnsureStringWithPrefix(c.Command, "--advertise-address=", address)
	}
	return nil
}
