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

	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

func ensureEnvVars(c *corev1.Container) {
	c.Env = controlplane.EnsureEnvVarWithName(c.Env,
		corev1.EnvVar{
			Name: "AWS_ACCESS_KEY_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key:                  aws.AccessKeyID,
					LocalObjectReference: corev1.LocalObjectReference{Name: common.CloudProviderSecretName},
				},
			},
		},
	)
	c.Env = controlplane.EnsureEnvVarWithName(c.Env,
		corev1.EnvVar{
			Name: "AWS_SECRET_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key:                  aws.SecretAccessKey,
					LocalObjectReference: corev1.LocalObjectReference{Name: common.CloudProviderSecretName},
				},
			},
		},
	)
}

func ensureVolumeMounts(c *corev1.Container) {
	c.VolumeMounts = controlplane.EnsureVolumeMountWithName(c.VolumeMounts,
		corev1.VolumeMount{
			Name:      aws.CloudProviderConfigName,
			MountPath: "/etc/kubernetes/cloudprovider",
		},
	)
	c.VolumeMounts = controlplane.EnsureVolumeMountWithName(c.VolumeMounts,
		corev1.VolumeMount{
			Name:      common.CloudProviderSecretName,
			MountPath: "/srv/cloudprovider",
		},
	)
}

func ensureVolumes(ps *corev1.PodSpec) {
	ps.Volumes = controlplane.EnsureVolumeWithName(ps.Volumes,
		corev1.Volume{
			Name: aws.CloudProviderConfigName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: aws.CloudProviderConfigName},
				},
			},
		},
	)
	ps.Volumes = controlplane.EnsureVolumeWithName(ps.Volumes,
		corev1.Volume{
			Name: common.CloudProviderSecretName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					// TODO Use constant from github.com/gardener/gardener/pkg/apis/core/v1alpha1 when available
					// See https://github.com/gardener/gardener/pull/930
					SecretName: common.CloudProviderSecretName,
				},
			},
		},
	)
}
