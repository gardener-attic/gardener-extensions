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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// newMutator creates a new controlplaneexposure mutator.
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
	case *corev1.Service:
		switch x.Name {
		case "kube-apiserver":
			return mutateKubeAPIServerService(x)
		}
	case *appsv1.Deployment:
		switch x.Name {
		case common.KubeAPIServerDeploymentName:
			return mutateKubeAPIServerDeployment(x)
		}
	}
	return nil
}

func mutateKubeAPIServerService(svc *corev1.Service) error {
	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}
	svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout"] = "3600"
	svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-backend-protocol"] = "ssl"
	svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-ssl-ports"] = "443"
	svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-healthcheck-timeout"] = "5"
	svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-healthcheck-interval"] = "30"
	svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-healthcheck-healthy-threshold"] = "2"
	svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-healthcheck-unhealthy-threshold"] = "2"
	svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-ssl-negotiation-policy"] = "ELBSecurityPolicy-TLS-1-2-2017-01"
	return nil
}

func mutateKubeAPIServerDeployment(dep *appsv1.Deployment) error {
	if c := controlplane.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-apiserver"); c != nil {
		c.Command = controlplane.EnsureStringWithPrefix(c.Command, "--endpoint-reconciler-type=", "none")
	}
	return nil
}
