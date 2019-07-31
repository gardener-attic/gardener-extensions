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

package shoot

import (
	"context"

	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/imagevector"
	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/packet"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	metabotInitContainerName = "metabot"
	metabotVolumeName        = "shared-init-config"
	metabotVolumeMountPath   = "/init-config"
)

func (m *mutator) mutateVPNShootDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	metabotImage, err := imagevector.ImageVector().FindImage(packet.MetabotImageName)
	if err != nil {
		return err
	}

	template := &deployment.Spec.Template
	ps := &template.Spec

	for _, initContainer := range ps.InitContainers {
		if initContainer.Name == metabotInitContainerName {
			return nil
		}
	}

	volumeMount := corev1.VolumeMount{
		Name:      metabotVolumeName,
		MountPath: metabotVolumeMountPath,
	}

	ps.InitContainers = append(ps.InitContainers, corev1.Container{
		Name:  metabotInitContainerName,
		Image: metabotImage.String(),
		Args: []string{
			"ip",
			"4",
			"private",
			"parent",
			"network",
		},
		VolumeMounts: []corev1.VolumeMount{volumeMount},
	})

	if c := extensionswebhook.ContainerWithName(ps.Containers, "vpn-shoot"); c != nil {
		c.VolumeMounts = extensionswebhook.EnsureVolumeMountWithName(c.VolumeMounts, volumeMount)
	}
	ps.Volumes = extensionswebhook.EnsureVolumeWithName(ps.Volumes, corev1.Volume{
		Name: metabotVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	return nil
}
