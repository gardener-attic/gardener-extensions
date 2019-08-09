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

	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/packet"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/genericmutator"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// NewEnsurer creates a new controlplaneexposure ensurer.
func NewEnsurer(imageVector imagevector.ImageVector, logger logr.Logger) genericmutator.Ensurer {
	return &ensurer{
		imageVector: imageVector,
		logger:      logger.WithName("packet-controlplanebackup-ensurer"),
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer
	imageVector imagevector.ImageVector
	logger      logr.Logger
}

// EnsureETCDStatefulSet ensures that the etcd stateful sets conform to the provider requirements.
func (e *ensurer) EnsureETCDStatefulSet(ctx context.Context, ss *appsv1.StatefulSet, cluster *extensionscontroller.Cluster) error {
	return e.ensureContainers(&ss.Spec.Template.Spec, ss.Name, cluster)
}

func (e *ensurer) ensureContainers(ps *corev1.PodSpec, name string, cluster *extensionscontroller.Cluster) error {
	c, err := e.getBackupRestoreContainer(name, cluster)
	if err != nil {
		return err
	}
	ps.Containers = extensionswebhook.EnsureContainerWithName(ps.Containers, *c)
	return nil
}

func (e *ensurer) getBackupRestoreContainer(name string, cluster *extensionscontroller.Cluster) (*corev1.Container, error) {
	// Find etcd-backup-restore image
	// TODO Get seed version from clientset when it's possible to inject it
	image, err := e.imageVector.FindImage(packet.ETCDBackupRestoreImageName, imagevector.TargetVersion(cluster.Shoot.Spec.Kubernetes.Version))
	if err != nil {
		return nil, errors.Wrapf(err, "could not find image %s", packet.ETCDBackupRestoreImageName)
	}

	// Determine volume claim template name
	// It is only specified for the etcd-main stateful set (backup is enabled)
	var (
		volumeClaimTemplateName = name
	)
	if name == gardencorev1alpha1.StatefulSetNameETCDMain {
		volumeClaimTemplateName = controlplane.EtcdMainVolumeClaimTemplateName
	}

	return controlplane.GetBackupRestoreContainer(name, volumeClaimTemplateName, "", "", "", image.String(), nil, nil, nil), nil
}
