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

package controlplanebackup

import (
	"context"

	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/config"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/genericmutator"
	v1alpha1constants "github.com/gardener/gardener/pkg/apis/core/v1alpha1/constants"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewEnsurer creates a new controlplaneexposure ensurer.
func NewEnsurer(etcdBackup *config.ETCDBackup, imageVector imagevector.ImageVector, logger logr.Logger) genericmutator.Ensurer {
	return &ensurer{
		etcdBackup:  etcdBackup,
		imageVector: imageVector,
		logger:      logger.WithName("alicloud-controlplanebackup-ensurer"),
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer
	etcdBackup  *config.ETCDBackup
	imageVector imagevector.ImageVector
	client      client.Client
	logger      logr.Logger
}

// InjectClient injects the given client into the ensurer.
func (e *ensurer) InjectClient(client client.Client) error {
	e.client = client
	return nil
}

// EnsureETCDStatefulSet ensures that the etcd stateful sets conform to the provider requirements.
func (e *ensurer) EnsureETCDStatefulSet(ctx context.Context, ss *appsv1.StatefulSet, cluster *extensionscontroller.Cluster) error {
	if err := e.ensureContainers(&ss.Spec.Template.Spec, ss.Name, cluster); err != nil {
		return err
	}
	return e.ensureChecksumAnnotations(ctx, &ss.Spec.Template, ss.Namespace, ss.Name, !extensionscontroller.IsSeedBackupNil(cluster))
}

func (e *ensurer) ensureContainers(ps *corev1.PodSpec, name string, cluster *extensionscontroller.Cluster) error {
	c, err := e.getBackupRestoreContainer(name, cluster)
	if err != nil {
		return err
	}
	ps.Containers = extensionswebhook.EnsureContainerWithName(ps.Containers, *c)
	return nil
}

func (e *ensurer) ensureChecksumAnnotations(ctx context.Context, template *corev1.PodTemplateSpec, namespace, name string, backupConfigured bool) error {
	if name == v1alpha1constants.StatefulSetNameETCDMain && backupConfigured {
		return controlplane.EnsureSecretChecksumAnnotation(ctx, template, e.client, namespace, alicloud.BackupSecretName)
	}
	return nil
}

func (e *ensurer) getBackupRestoreContainer(name string, cluster *extensionscontroller.Cluster) (*corev1.Container, error) {
	// Find etcd-backup-restore image
	// TODO Get seed version from clientset when it's possible to inject it
	image, err := e.imageVector.FindImage(alicloud.ETCDBackupRestoreImageName, imagevector.TargetVersion(extensionscontroller.GetKubernetesVersion(cluster)))
	if err != nil {
		return nil, errors.Wrapf(err, "could not find image %s", alicloud.ETCDBackupRestoreImageName)
	}

	const (
		defaultSchedule = "0 */24 * * *"
	)

	// Determine provider and container env variables
	// They are only specified for the etcd-main stateful set (backup is enabled)
	var (
		provider                string
		prefix                  string
		env                     []corev1.EnvVar
		volumeClaimTemplateName = name
	)
	if name == v1alpha1constants.StatefulSetNameETCDMain {
		if extensionscontroller.IsSeedBackupNil(cluster) {
			e.logger.Info("Backup profile is not configured; backup will not be taken for etcd-main")
		} else {
			prefix = common.GenerateBackupEntryName(extensionscontroller.GetTechnicalID(cluster), extensionscontroller.GetUID(cluster))
			provider = alicloud.StorageProviderName
			env = []corev1.EnvVar{
				{
					Name: "STORAGE_CONTAINER",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key:                  alicloud.BucketName,
							LocalObjectReference: corev1.LocalObjectReference{Name: alicloud.BackupSecretName},
						},
					},
				},
				{
					Name: "ALICLOUD_ENDPOINT",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key:                  alicloud.StorageEndpoint,
							LocalObjectReference: corev1.LocalObjectReference{Name: alicloud.BackupSecretName},
						},
					},
				},
				{
					Name: "ALICLOUD_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key:                  alicloud.AccessKeyID,
							LocalObjectReference: corev1.LocalObjectReference{Name: alicloud.BackupSecretName},
						},
					},
				},
				{
					Name: "ALICLOUD_ACCESS_KEY_SECRET",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key:                  alicloud.AccessKeySecret,
							LocalObjectReference: corev1.LocalObjectReference{Name: alicloud.BackupSecretName},
						},
					},
				},
			}
		}
		volumeClaimTemplateName = controlplane.EtcdMainVolumeClaimTemplateName
	}

	// Determine schedule
	var schedule = defaultSchedule
	if e.etcdBackup.Schedule != nil {
		schedule = defaultSchedule
	}

	return controlplane.GetBackupRestoreContainer(name, volumeClaimTemplateName, schedule, provider, prefix, image.String(), nil, env, nil), nil
}
