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

	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/packet"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/genericmutator"

	"github.com/coreos/go-systemd/unit"
	v1alpha1constants "github.com/gardener/gardener/pkg/apis/core/v1alpha1/constants"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewEnsurer creates a new controlplane ensurer.
func NewEnsurer(logger logr.Logger) genericmutator.Ensurer {
	return &ensurer{
		logger: logger.WithName("packet-controlplane-ensurer"),
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer
	client client.Client
	logger logr.Logger
}

// InjectClient injects the given client into the ensurer.
func (e *ensurer) InjectClient(client client.Client) error {
	e.client = client
	return nil
}

// EnsureKubeAPIServerDeployment ensures that the kube-apiserver deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeAPIServerDeployment(ctx context.Context, dep *appsv1.Deployment) error {
	ps := &dep.Spec.Template.Spec
	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-apiserver"); c != nil {
		ensureKubeAPIServerCommandLineArgs(c)
		ensureEnvVars(c)
	}
	return controlplane.EnsureSecretChecksumAnnotation(ctx, &dep.Spec.Template, e.client, dep.Namespace, v1alpha1constants.SecretNameCloudProvider)
}

// EnsureKubeControllerManagerDeployment ensures that the kube-controller-manager deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeControllerManagerDeployment(ctx context.Context, dep *appsv1.Deployment) error {
	ps := &dep.Spec.Template.Spec
	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-controller-manager"); c != nil {
		ensureKubeControllerManagerCommandLineArgs(c)
	}
	return nil
}

func ensureKubeAPIServerCommandLineArgs(c *corev1.Container) {
	// Ensure CSI-related admission plugins
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--enable-admission-plugins=",
		"PersistentVolumeLabel", ",")
	c.Command = extensionswebhook.EnsureStringWithPrefixContains(c.Command, "--disable-admission-plugins=",
		"PersistentVolumeLabel", ",")

	// Ensure CSI-related feature gates
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"VolumeSnapshotDataSource=false", ",")
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"CSINodeInfo=false", ",")
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"CSIDriverRegistry=false", ",")
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"KubeletPluginsWatcher=false", ",")
	c.Command = extensionswebhook.EnsureStringWithPrefixContains(c.Command, "--feature-gates=",
		"VolumeSnapshotDataSource=true", ",")
	c.Command = extensionswebhook.EnsureStringWithPrefixContains(c.Command, "--feature-gates=",
		"CSINodeInfo=true", ",")
	c.Command = extensionswebhook.EnsureStringWithPrefixContains(c.Command, "--feature-gates=",
		"CSIDriverRegistry=true", ",")
}

func ensureKubeControllerManagerCommandLineArgs(c *corev1.Container) {
	c.Command = extensionswebhook.EnsureStringWithPrefix(c.Command, "--cloud-provider=", "external")
}

var (
	credentialsEnvVar = corev1.EnvVar{
		Name: "PACKET_API_KEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: packet.APIToken,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: v1alpha1constants.SecretNameCloudProvider,
				},
			},
		},
	}
)

func ensureEnvVars(c *corev1.Container) {
	c.Env = extensionswebhook.EnsureEnvVarWithName(c.Env, credentialsEnvVar)
}

// EnsureKubeletServiceUnitOptions ensures that the kubelet.service unit options conform to the provider requirements.
func (e *ensurer) EnsureKubeletServiceUnitOptions(ctx context.Context, opts []*unit.UnitOption) ([]*unit.UnitOption, error) {
	if opt := extensionswebhook.UnitOptionWithSectionAndName(opts, "Service", "ExecStart"); opt != nil {
		command := extensionswebhook.DeserializeCommandLine(opt.Value)
		command = ensureKubeletCommandLineArgs(command)
		opt.Value = extensionswebhook.SerializeCommandLine(command, 1, " \\\n    ")
	}
	return opts, nil
}

func ensureKubeletCommandLineArgs(command []string) []string {
	command = extensionswebhook.EnsureStringWithPrefix(command, "--cloud-provider=", "external")
	command = extensionswebhook.EnsureStringWithPrefix(command, "--enable-controller-attach-detach=", "true")
	return command
}

// EnsureKubeletConfiguration ensures that the kubelet configuration conforms to the provider requirements.
func (e *ensurer) EnsureKubeletConfiguration(ctx context.Context, kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration) error {
	// Ensure CSI-related feature gates
	if kubeletConfig.FeatureGates == nil {
		kubeletConfig.FeatureGates = make(map[string]bool)
	}
	kubeletConfig.FeatureGates["VolumeSnapshotDataSource"] = true
	kubeletConfig.FeatureGates["CSINodeInfo"] = true
	kubeletConfig.FeatureGates["CSIDriverRegistry"] = true
	return nil
}
