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

package genericmutator

import (
	"context"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"

	"github.com/coreos/go-systemd/unit"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// TODO add it in github.com/gardener/gardener/pkg/operation/common
// KubeAPIServerNetworkPolicyName is the name of the kube-apiserver networkpolicy.
const KubeAPIServerNetworkPolicyName = "kube-apiserver-default"

// Ensurer ensures that various standard Kubernets controlplane objects conform to the provider requirements.
// If they don't initially, they are mutated accordingly.
type Ensurer interface {
	// EnsureKubeAPIServerService ensures that the kube-apiserver service conforms to the provider requirements.
	EnsureKubeAPIServerService(context.Context, *corev1.Service) error
	// EnsureKubeAPIServerDeployment ensures that the kube-apiserver deployment conforms to the provider requirements.
	EnsureKubeAPIServerDeployment(context.Context, *appsv1.Deployment) error
	// EnsureKubeAPIServerNetworkPolicy ensures that the kube-apiserver network policy conforms to the provider requirements.
	EnsureKubeAPIServerNetworkPolicy(context.Context, *networkingv1.NetworkPolicy) error
	// EnsureKubeControllerManagerDeployment ensures that the kube-controller-manager deployment conforms to the provider requirements.
	EnsureKubeControllerManagerDeployment(context.Context, *appsv1.Deployment) error
	// EnsureKubeSchedulerDeployment ensures that the kube-scheduler deployment conforms to the provider requirements.
	EnsureKubeSchedulerDeployment(context.Context, *appsv1.Deployment) error
	// EnsureETCDStatefulSet ensures that the etcd stateful sets conform to the provider requirements.
	EnsureETCDStatefulSet(context.Context, *appsv1.StatefulSet, *extensionscontroller.Cluster) error
	// EnsureKubeletServiceUnitOptions ensures that the kubelet.service unit options conform to the provider requirements.
	EnsureKubeletServiceUnitOptions(context.Context, []*unit.UnitOption) ([]*unit.UnitOption, error)
	// EnsureKubeletConfiguration ensures that the kubelet configuration conforms to the provider requirements.
	EnsureKubeletConfiguration(context.Context, *kubeletconfigv1beta1.KubeletConfiguration) error
}

// NewMutator creates a new controlplane mutator.
func NewMutator(ensurer Ensurer, unitSerializer controlplane.UnitSerializer, kubeletConfigCodec controlplane.KubeletConfigCodec, logger logr.Logger) controlplane.Mutator {
	return &mutator{
		ensurer:            ensurer,
		unitSerializer:     unitSerializer,
		kubeletConfigCodec: kubeletConfigCodec,
		logger:             logger.WithName("mutator"),
	}
}

type mutator struct {
	client             client.Client
	ensurer            Ensurer
	unitSerializer     controlplane.UnitSerializer
	kubeletConfigCodec controlplane.KubeletConfigCodec
	logger             logr.Logger
}

// InjectClient injects the given client into the ensurer.
// TODO Replace this with the more generic InjectFunc when controller runtime supports it
func (m *mutator) InjectClient(client client.Client) error {
	m.client = client
	if _, err := inject.ClientInto(client, m.ensurer); err != nil {
		return errors.Wrap(err, "could not inject the client into the ensurer")
	}
	return nil
}

// Mutate validates and if needed mutates the given object.
func (m *mutator) Mutate(ctx context.Context, obj runtime.Object) error {
	switch x := obj.(type) {
	case *corev1.Service:
		switch x.Name {
		case common.KubeAPIServerDeploymentName:
			return m.ensurer.EnsureKubeAPIServerService(ctx, x)
		}
	case *appsv1.Deployment:
		switch x.Name {
		case common.KubeAPIServerDeploymentName:
			return m.ensurer.EnsureKubeAPIServerDeployment(ctx, x)
		case common.KubeControllerManagerDeploymentName:
			return m.ensurer.EnsureKubeControllerManagerDeployment(ctx, x)
		case common.KubeSchedulerDeploymentName:
			return m.ensurer.EnsureKubeSchedulerDeployment(ctx, x)
		}
	case *appsv1.StatefulSet:
		switch x.Name {
		case common.EtcdMainStatefulSetName, common.EtcdEventsStatefulSetName:
			// Get cluster info
			cluster, err := extensionscontroller.GetCluster(ctx, m.client, x.Namespace)
			if err != nil {
				return errors.Wrapf(err, "could not get cluster for namespace '%s'", x.Namespace)
			}

			return m.ensurer.EnsureETCDStatefulSet(ctx, x, cluster)
		}
	case *extensionsv1alpha1.OperatingSystemConfig:
		if x.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeReconcile {
			return m.mutateOperatingSystemConfig(ctx, x)
		}
	case *networkingv1.NetworkPolicy:
		if x.Name == KubeAPIServerNetworkPolicyName {
			return m.ensurer.EnsureKubeAPIServerNetworkPolicy(ctx, x)
		}
	}

	return nil
}

func (m *mutator) mutateOperatingSystemConfig(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig) error {
	// Mutate kubelet.service unit, if present
	if u := controlplane.UnitWithName(osc.Spec.Units, "kubelet.service"); u != nil && u.Content != nil {
		if err := m.ensureKubeletServiceUnitContent(ctx, u.Content); err != nil {
			return err
		}
	}

	// Mutate kubelet configuration file, if present
	if f := controlplane.FileWithPath(osc.Spec.Files, "/var/lib/kubelet/config/kubelet"); f != nil && f.Content.Inline != nil {
		if err := m.ensureKubeletConfigFileContent(ctx, f.Content.Inline); err != nil {
			return err
		}
	}

	return nil
}

func (m *mutator) ensureKubeletServiceUnitContent(ctx context.Context, content *string) error {
	var opts []*unit.UnitOption
	var err error

	// Deserialize unit options
	if opts, err = m.unitSerializer.Deserialize(*content); err != nil {
		return errors.Wrap(err, "could not deserialize kubelet.service unit content")
	}

	if opts, err = m.ensurer.EnsureKubeletServiceUnitOptions(ctx, opts); err != nil {
		return err
	}

	// Serialize unit options
	if *content, err = m.unitSerializer.Serialize(opts); err != nil {
		return errors.Wrap(err, "could not serialize kubelet.service unit options")
	}

	return nil
}

func (m *mutator) ensureKubeletConfigFileContent(ctx context.Context, fci *extensionsv1alpha1.FileContentInline) error {
	var kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration
	var err error

	// Decode kubelet configuration from inline content
	if kubeletConfig, err = m.kubeletConfigCodec.Decode(fci); err != nil {
		return errors.Wrap(err, "could not decode kubelet configuration")
	}

	if err = m.ensurer.EnsureKubeletConfiguration(ctx, kubeletConfig); err != nil {
		return err
	}

	// Encode kubelet configuration into inline content
	var newFCI *extensionsv1alpha1.FileContentInline
	if newFCI, err = m.kubeletConfigCodec.Encode(kubeletConfig, fci.Encoding); err != nil {
		return errors.Wrap(err, "could not encode kubelet configuration")
	}
	*fci = *newFCI

	return nil
}
