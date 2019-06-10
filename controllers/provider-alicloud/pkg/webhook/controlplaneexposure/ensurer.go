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
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/config"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/genericmutator"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewEnsurer creates a new controlplaneexposure ensurer.
func NewEnsurer(etcdStorage *config.ETCDStorage, logger logr.Logger) genericmutator.Ensurer {
	return &ensurer{
		etcdStorage: etcdStorage,
		logger:      logger.WithName("alicloud-controlplaneexposure-ensurer"),
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer
	etcdStorage *config.ETCDStorage
	client      client.Client
	logger      logr.Logger
}

// InjectClient injects the given client into the ensurer.
func (m *ensurer) InjectClient(client client.Client) error {
	m.client = client
	return nil
}

const (
	AlicloudCpCidr    = "100.100.100.200/32"
	DefaultSeedCpCidr = "169.254.169.254/32"

	notFound = -1
)

// EnsureKubeAPIServerNetworkPolicy ensures that the kube-apiserver network policy conforms to the provider requirements.
func (e *ensurer) EnsureKubeAPIServerNetworkPolicy(ctx context.Context, np *networkingv1.NetworkPolicy) error {
	egressRules := np.Spec.Egress
	for _, egress := range egressRules {
		if len(egress.To) == 0 {
			continue
		}
		// TODO should I check the rest of the values inside the IPBlock.Except list?
		for _, to := range egress.To {
			if to.IPBlock != nil {
				cidrs := to.IPBlock.Except
				i := find(cidrs, DefaultSeedCpCidr)

				if i != notFound {
					// Remove default CIDR
					cidrs = append(cidrs[:i], cidrs[i+1:]...)
				}
				// TODO should we throw an error if the default IP is not found?
				cidrs = append(cidrs, AlicloudCpCidr)
				to.IPBlock.Except = cidrs
			}
		}
	}
	return nil
}

// EnsureKubeAPIServerDeployment ensures that the kube-apiserver deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeAPIServerDeployment(ctx context.Context, dep *appsv1.Deployment) error {
	// Get load balancer address of the kube-apiserver service
	address, err := controlplane.GetLoadBalancerIngress(ctx, e.client, dep.Namespace, common.KubeAPIServerDeploymentName)
	if err != nil {
		return errors.Wrap(err, "could not get kube-apiserver service load balancer address")
	}

	if c := controlplane.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-apiserver"); c != nil {
		c.Command = controlplane.EnsureStringWithPrefix(c.Command, "--advertise-address=", address)
		c.Command = controlplane.EnsureStringWithPrefix(c.Command, "--external-hostname=", address)
	}
	return nil
}

// EnsureETCDStatefulSet ensures that the etcd stateful sets conform to the provider requirements.
func (e *ensurer) EnsureETCDStatefulSet(ctx context.Context, ss *appsv1.StatefulSet, cluster *extensionscontroller.Cluster) error {
	e.ensureVolumeClaimTemplates(&ss.Spec, ss.Name)
	return nil
}

func (e *ensurer) ensureVolumeClaimTemplates(spec *appsv1.StatefulSetSpec, name string) {
	t := e.getVolumeClaimTemplate(name)
	spec.VolumeClaimTemplates = controlplane.EnsurePVCWithName(spec.VolumeClaimTemplates, *t)
}

func (e *ensurer) getVolumeClaimTemplate(name string) *corev1.PersistentVolumeClaim {
	var etcdStorage config.ETCDStorage
	if name == common.EtcdMainStatefulSetName {
		etcdStorage = *e.etcdStorage
	}
	return controlplane.GetETCDVolumeClaimTemplate(name, etcdStorage.ClassName, etcdStorage.Capacity)
}

func find(list []string, element string) int {
	if len(list) == 0 {
		return notFound
	}
	for i, str := range list {
		if str == element {
			return i
		}
	}
	return notFound
}
