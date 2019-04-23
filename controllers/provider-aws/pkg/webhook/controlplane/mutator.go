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

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"

	"github.com/coreos/go-systemd/unit"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

// newMutator creates a new controlplane mutator.
func newMutator(unitSerializer controlplane.UnitSerializer, kubeletConfigCodec controlplane.KubeletConfigCodec, logger logr.Logger) *mutator {
	return &mutator{
		unitSerializer:     unitSerializer,
		kubeletConfigCodec: kubeletConfigCodec,
		logger:             logger.WithName("mutator"),
	}
}

type mutator struct {
	unitSerializer     controlplane.UnitSerializer
	kubeletConfigCodec controlplane.KubeletConfigCodec
	logger             logr.Logger
}

// Mutate validates and if needed mutates the given object.
func (m *mutator) Mutate(ctx context.Context, obj runtime.Object) error {
	switch x := obj.(type) {
	case *appsv1.Deployment:
		switch x.Name {
		case common.KubeAPIServerDeploymentName:
			return mutateKubeAPIServerDeployment(x)
		case common.KubeControllerManagerDeploymentName:
			return mutateKubeControllerManagerDeployment(x)
		}
	case *extensionsv1alpha1.OperatingSystemConfig:
		if x.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeReconcile {
			return m.mutateOperatingSystemConfig(x)
		}
	}
	return nil
}

func mutateKubeAPIServerDeployment(dep *appsv1.Deployment) error {
	ps := &dep.Spec.Template.Spec
	if c := controlplane.ContainerWithName(ps.Containers, "kube-apiserver"); c != nil {
		ensureKubeAPIServerCommandLineArgs(c)
		ensureEnvVars(c)
		ensureVolumeMounts(c)
	}
	ensureVolumes(ps)
	return nil
}

func mutateKubeControllerManagerDeployment(dep *appsv1.Deployment) error {
	ps := &dep.Spec.Template.Spec
	if c := controlplane.ContainerWithName(ps.Containers, "kube-controller-manager"); c != nil {
		ensureKubeControllerManagerCommandLineArgs(c)
		ensureEnvVars(c)
		ensureVolumeMounts(c)
	}
	ensureVolumes(ps)
	return nil
}

func ensureKubeAPIServerCommandLineArgs(c *corev1.Container) {
	c.Command = controlplane.EnsureStringWithPrefix(c.Command, "--cloud-provider=", "aws")
	c.Command = controlplane.EnsureStringWithPrefix(c.Command, "--cloud-config=",
		"/etc/kubernetes/cloudprovider/cloudprovider.conf")
	c.Command = controlplane.EnsureStringWithPrefixContains(c.Command, "--enable-admission-plugins=",
		"PersistentVolumeLabel", ",")
	c.Command = controlplane.EnsureNoStringWithPrefixContains(c.Command, "--disable-admission-plugins=",
		"PersistentVolumeLabel", ",")
}

func ensureKubeControllerManagerCommandLineArgs(c *corev1.Container) {
	c.Command = controlplane.EnsureStringWithPrefix(c.Command, "--cloud-provider=", "external")
	c.Command = controlplane.EnsureStringWithPrefix(c.Command, "--cloud-config=",
		"/etc/kubernetes/cloudprovider/cloudprovider.conf")
	c.Command = controlplane.EnsureStringWithPrefix(c.Command, "--external-cloud-volume-plugin=", "aws")
}

var (
	accessKeyIDEnvVar = corev1.EnvVar{
		Name: "AWS_ACCESS_KEY_ID",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key:                  aws.AccessKeyID,
				LocalObjectReference: corev1.LocalObjectReference{Name: common.CloudProviderSecretName},
			},
		},
	}
	secretAccessKeyEnvVar = corev1.EnvVar{
		Name: "AWS_SECRET_ACCESS_KEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key:                  aws.SecretAccessKey,
				LocalObjectReference: corev1.LocalObjectReference{Name: common.CloudProviderSecretName},
			},
		},
	}
)

func ensureEnvVars(c *corev1.Container) {
	c.Env = controlplane.EnsureEnvVarWithName(c.Env, accessKeyIDEnvVar)
	c.Env = controlplane.EnsureEnvVarWithName(c.Env, secretAccessKeyEnvVar)
}

var (
	cloudProviderConfigVolumeMount = corev1.VolumeMount{
		Name:      aws.CloudProviderConfigName,
		MountPath: "/etc/kubernetes/cloudprovider",
	}
	cloudProviderSecretVolumeMount = corev1.VolumeMount{
		Name:      common.CloudProviderSecretName,
		MountPath: "/srv/cloudprovider",
	}

	cloudProviderConfigVolume = corev1.Volume{
		Name: aws.CloudProviderConfigName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: aws.CloudProviderConfigName},
			},
		},
	}
	cloudProviderSecretVolume = corev1.Volume{
		Name: common.CloudProviderSecretName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				// TODO Use constant from github.com/gardener/gardener/pkg/apis/core/v1alpha1 when available
				// See https://github.com/gardener/gardener/pull/930
				SecretName: common.CloudProviderSecretName,
			},
		},
	}
)

func ensureVolumeMounts(c *corev1.Container) {
	c.VolumeMounts = controlplane.EnsureVolumeMountWithName(c.VolumeMounts, cloudProviderConfigVolumeMount)
	c.VolumeMounts = controlplane.EnsureVolumeMountWithName(c.VolumeMounts, cloudProviderSecretVolumeMount)
}

func ensureVolumes(ps *corev1.PodSpec) {
	ps.Volumes = controlplane.EnsureVolumeWithName(ps.Volumes, cloudProviderConfigVolume)
	ps.Volumes = controlplane.EnsureVolumeWithName(ps.Volumes, cloudProviderSecretVolume)
}

func (m *mutator) mutateOperatingSystemConfig(osc *extensionsv1alpha1.OperatingSystemConfig) error {
	// Mutate kubelet.service unit, if present
	if u := controlplane.UnitWithName(osc.Spec.Units, "kubelet.service"); u != nil && u.Content != nil {
		if err := m.ensureKubeletServiceUnitContent(u.Content); err != nil {
			return err
		}
	}

	// Mutate kubelet configuration file, if present
	if f := controlplane.FileWithPath(osc.Spec.Files, "/var/lib/kubelet/config/kubelet"); f != nil && f.Content.Inline != nil {
		if err := m.ensureKubeletConfigFileContent(f.Content.Inline); err != nil {
			return err
		}
	}

	return nil
}

func (m *mutator) ensureKubeletServiceUnitContent(content *string) error {
	var opts []*unit.UnitOption
	var err error

	// Deserialize unit options
	if opts, err = m.unitSerializer.Deserialize(*content); err != nil {
		return errors.Wrap(err, "could not deserialize kubelet.service unit content")
	}

	ensureKubeletServiceUnitOptions(opts)

	// Serialize unit options
	if *content, err = m.unitSerializer.Serialize(opts); err != nil {
		return errors.Wrap(err, "could not serialize kubelet.service unit options")
	}

	return nil
}

func (m *mutator) ensureKubeletConfigFileContent(fci *extensionsv1alpha1.FileContentInline) error {
	var kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration
	var err error

	// Decode kubelet configuration from inline content
	if kubeletConfig, err = m.kubeletConfigCodec.Decode(fci); err != nil {
		return errors.Wrap(err, "could not decode kubelet configuration")
	}

	ensureKubeletConfiguration(kubeletConfig)

	// Encode kubelet configuration into inline content
	var newFCI *extensionsv1alpha1.FileContentInline
	if newFCI, err = m.kubeletConfigCodec.Encode(kubeletConfig, fci.Encoding); err != nil {
		return errors.Wrap(err, "could not encode kubelet configuration")
	}
	*fci = *newFCI

	return nil
}

func ensureKubeletServiceUnitOptions(opts []*unit.UnitOption) {
	if opt := controlplane.UnitOptionWithSectionAndName(opts, "Service", "ExecStart"); opt != nil {
		command := controlplane.DeserializeCommandLine(opt.Value)
		command = ensureKubeletCommandLineArgs(command)
		opt.Value = controlplane.SerializeCommandLine(command, 1, " \\\n    ")
	}
}

func ensureKubeletCommandLineArgs(command []string) []string {
	command = controlplane.EnsureStringWithPrefix(command, "--cloud-provider=", "aws")
	return command
}

func ensureKubeletConfiguration(kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration) {
	// Make sure CSI-related feature gates are not enabled
	// TODO Leaving these enabled shouldn't do any harm, perhaps remove this code when properly tested?
	delete(kubeletConfig.FeatureGates, "VolumeSnapshotDataSource")
	delete(kubeletConfig.FeatureGates, "CSINodeInfo")
	delete(kubeletConfig.FeatureGates, "CSIDriverRegistry")
}
