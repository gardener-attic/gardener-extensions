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
	"bytes"
	"context"
	"fmt"
	"regexp"

	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/gcp"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal"
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
		logger: logger.WithName("gcp-controlplane-ensurer"),
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
func (e *ensurer) EnsureKubeAPIServerDeployment(ctx genericmutator.EnsurerContext, dep *appsv1.Deployment) error {
	template := &dep.Spec.Template
	ps := &template.Spec
	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-apiserver"); c != nil {
		ensureKubeAPIServerCommandLineArgs(c)
		ensureEnvVars(c)
		ensureVolumeMounts(c)
	}
	ensureVolumes(ps)
	return e.ensureChecksumAnnotations(ctx.GetContext(), &dep.Spec.Template, dep.Namespace)
}

// EnsureKubeControllerManagerDeployment ensures that the kube-controller-manager deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeControllerManagerDeployment(ctx genericmutator.EnsurerContext, dep *appsv1.Deployment) error {
	template := &dep.Spec.Template
	ps := &template.Spec
	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-controller-manager"); c != nil {
		ensureKubeControllerManagerCommandLineArgs(c)
		ensureEnvVars(c)
		ensureVolumeMounts(c)
	}
	ensureKubeControllerManagerAnnotations(template)
	ensureVolumes(ps)
	return e.ensureChecksumAnnotations(ctx.GetContext(), &dep.Spec.Template, dep.Namespace)
}

func ensureKubeAPIServerCommandLineArgs(c *corev1.Container) {
	c.Command = extensionswebhook.EnsureStringWithPrefix(c.Command, "--cloud-provider=", "gce")
	c.Command = extensionswebhook.EnsureStringWithPrefix(c.Command, "--cloud-config=",
		"/etc/kubernetes/cloudprovider/cloudprovider.conf")
	c.Command = extensionswebhook.EnsureStringWithPrefixContains(c.Command, "--enable-admission-plugins=",
		"PersistentVolumeLabel", ",")
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--disable-admission-plugins=",
		"PersistentVolumeLabel", ",")
}

func ensureKubeControllerManagerCommandLineArgs(c *corev1.Container) {
	c.Command = extensionswebhook.EnsureStringWithPrefix(c.Command, "--cloud-provider=", "external")
	c.Command = extensionswebhook.EnsureStringWithPrefix(c.Command, "--cloud-config=",
		"/etc/kubernetes/cloudprovider/cloudprovider.conf")
	c.Command = extensionswebhook.EnsureStringWithPrefix(c.Command, "--external-cloud-volume-plugin=", "gce")
}

func ensureKubeControllerManagerAnnotations(t *corev1.PodTemplateSpec) {
	t.Labels = extensionswebhook.EnsureAnnotationOrLabel(t.Labels, v1alpha1constants.LabelNetworkPolicyToPublicNetworks, v1alpha1constants.LabelNetworkPolicyAllowed)
	t.Labels = extensionswebhook.EnsureAnnotationOrLabel(t.Labels, v1alpha1constants.LabelNetworkPolicyToPrivateNetworks, v1alpha1constants.LabelNetworkPolicyAllowed)
	t.Labels = extensionswebhook.EnsureAnnotationOrLabel(t.Labels, v1alpha1constants.LabelNetworkPolicyToBlockedCIDRs, v1alpha1constants.LabelNetworkPolicyAllowed)
}

var (
	credentialsEnvVar = corev1.EnvVar{
		Name:  "GOOGLE_APPLICATION_CREDENTIALS",
		Value: fmt.Sprintf("/srv/cloudprovider/%s", gcp.ServiceAccountJSONField),
	}
)

func ensureEnvVars(c *corev1.Container) {
	c.Env = extensionswebhook.EnsureEnvVarWithName(c.Env, credentialsEnvVar)
}

var (
	cloudProviderConfigVolumeMount = corev1.VolumeMount{
		Name:      internal.CloudProviderConfigName,
		MountPath: "/etc/kubernetes/cloudprovider",
	}
	cloudProviderSecretVolumeMount = corev1.VolumeMount{
		Name:      v1alpha1constants.SecretNameCloudProvider,
		MountPath: "/srv/cloudprovider",
	}

	cloudProviderConfigVolume = corev1.Volume{
		Name: internal.CloudProviderConfigName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: internal.CloudProviderConfigName},
			},
		},
	}
	cloudProviderSecretVolume = corev1.Volume{
		Name: v1alpha1constants.SecretNameCloudProvider,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: v1alpha1constants.SecretNameCloudProvider,
			},
		},
	}
)

func ensureVolumeMounts(c *corev1.Container) {
	c.VolumeMounts = extensionswebhook.EnsureVolumeMountWithName(c.VolumeMounts, cloudProviderConfigVolumeMount)
	c.VolumeMounts = extensionswebhook.EnsureVolumeMountWithName(c.VolumeMounts, cloudProviderSecretVolumeMount)
}

func ensureVolumes(ps *corev1.PodSpec) {
	ps.Volumes = extensionswebhook.EnsureVolumeWithName(ps.Volumes, cloudProviderConfigVolume)
	ps.Volumes = extensionswebhook.EnsureVolumeWithName(ps.Volumes, cloudProviderSecretVolume)
}

func (e *ensurer) ensureChecksumAnnotations(ctx context.Context, template *corev1.PodTemplateSpec, namespace string) error {
	if err := controlplane.EnsureSecretChecksumAnnotation(ctx, template, e.client, namespace, v1alpha1constants.SecretNameCloudProvider); err != nil {
		return err
	}
	return controlplane.EnsureConfigMapChecksumAnnotation(ctx, template, e.client, namespace, internal.CloudProviderConfigName)
}

// EnsureKubeletServiceUnitOptions ensures that the kubelet.service unit options conform to the provider requirements.
func (e *ensurer) EnsureKubeletServiceUnitOptions(ctx genericmutator.EnsurerContext, opts []*unit.UnitOption) ([]*unit.UnitOption, error) {
	if opt := extensionswebhook.UnitOptionWithSectionAndName(opts, "Service", "ExecStart"); opt != nil {
		command := extensionswebhook.DeserializeCommandLine(opt.Value)
		command = ensureKubeletCommandLineArgs(command)
		opt.Value = extensionswebhook.SerializeCommandLine(command, 1, " \\\n    ")
	}
	opts = extensionswebhook.EnsureUnitOption(opts, &unit.UnitOption{
		Section: "Service",
		Name:    "ExecStartPre",
		Value:   `/bin/sh -c 'hostnamectl set-hostname $(cat /etc/hostname | cut -d '.' -f 1)'`,
	})
	return opts, nil
}

func ensureKubeletCommandLineArgs(command []string) []string {
	command = extensionswebhook.EnsureStringWithPrefix(command, "--cloud-provider=", "gce")
	return command
}

// EnsureKubeletConfiguration ensures that the kubelet configuration conforms to the provider requirements.
func (e *ensurer) EnsureKubeletConfiguration(ctx genericmutator.EnsurerContext, kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration) error {
	// Make sure CSI-related feature gates are not enabled
	// TODO Leaving these enabled shouldn't do any harm, perhaps remove this code when properly tested?
	delete(kubeletConfig.FeatureGates, "VolumeSnapshotDataSource")
	delete(kubeletConfig.FeatureGates, "CSINodeInfo")
	delete(kubeletConfig.FeatureGates, "CSIDriverRegistry")
	return nil
}

var regexFindProperty = regexp.MustCompile("net.ipv4.ip_forward[[:space:]]*=[[:space:]]*([[:alnum:]]+)")

// EnsureKubernetesGeneralConfiguration ensures that the kubernetes general configuration conforms to the provider requirements.
func (e *ensurer) EnsureKubernetesGeneralConfiguration(ctx genericmutator.EnsurerContext, data *string) error {
	// If the needed property exists, ensure the correct value
	if regexFindProperty.MatchString(*data) {
		res := regexFindProperty.ReplaceAll([]byte(*data), []byte("net.ipv4.ip_forward = 1"))
		*data = string(res)
		return nil
	}

	// If the property do not exist, append it in the end of the string
	buf := bytes.Buffer{}
	buf.WriteString(*data)
	buf.WriteString("\n")
	buf.WriteString("# GCE specific settings\n")
	buf.WriteString("net.ipv4.ip_forward = 1")

	*data = buf.String()
	return nil
}
