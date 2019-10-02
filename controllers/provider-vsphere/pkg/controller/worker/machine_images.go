/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package worker

import (
	"context"
	"fmt"

	apisvsphere "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere"
	"github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere/helper"
	apisvspherehelper "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere/helper"
	vspherev1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere/v1alpha1"
	"github.com/gardener/gardener-extensions/pkg/util"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetMachineImages returns the used machine images for the `Worker` resource.
func (w *workerDelegate) GetMachineImages(ctx context.Context) (runtime.Object, error) {
	if w.machineImages == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return nil, err
		}
	}

	var (
		workerStatus = &apisvsphere.WorkerStatus{
			TypeMeta: metav1.TypeMeta{
				APIVersion: apisvsphere.SchemeGroupVersion.String(),
				Kind:       "WorkerStatus",
			},
			MachineImages: w.machineImages,
		}

		workerStatusV1alpha1 = &vspherev1alpha1.WorkerStatus{
			TypeMeta: metav1.TypeMeta{
				APIVersion: vspherev1alpha1.SchemeGroupVersion.String(),
				Kind:       "WorkerStatus",
			},
		}
	)

	if err := w.Scheme().Convert(workerStatus, workerStatusV1alpha1, nil); err != nil {
		return nil, err
	}

	return workerStatusV1alpha1, nil
}

func (w *workerDelegate) findMachineImage(name, version string) (string, string, error) {
	region := apisvspherehelper.FindRegion(w.cluster.Shoot.Spec.Region, w.profileConfig)
	if region != nil && len(region.MachineImages) != 0 {
		path, guestID, err := helper.FindImage(region.MachineImages, name, version)
		if err == nil {
			return path, guestID, nil
		}
	}

	if w.profileConfig != nil {
		path, guestID, err := helper.FindImage(w.profileConfig.MachineImages, name, version)
		if err == nil {
			return path, guestID, nil
		}
	}

	// Try to look up machine image in worker provider status as it was not found in componentconfig.
	if providerStatus := w.worker.Status.ProviderStatus; providerStatus != nil {
		workerStatus := &apisvsphere.WorkerStatus{}
		if _, _, err := w.Decoder().Decode(providerStatus.Raw, nil, workerStatus); err != nil {
			return "", "", errors.Wrapf(err, "could not decode worker status of worker '%s'", util.ObjectName(w.worker))
		}

		machineImage, err := apisvspherehelper.FindMachineImage(workerStatus.MachineImages, name, version)
		if err != nil {
			return "", "", errorMachineImageNotFound(name, version)
		}

		return machineImage.Path, machineImage.GuestID, nil
	}

	return "", "", errorMachineImageNotFound(name, version)
}

func errorMachineImageNotFound(name, version string) error {
	return fmt.Errorf("could not find machine image for %s/%s neither in componentconfig nor in worker status", name, version)
}

func appendMachineImage(machineImages []apisvsphere.MachineImage, machineImage apisvsphere.MachineImage) []apisvsphere.MachineImage {
	if _, err := apisvspherehelper.FindMachineImage(machineImages, machineImage.Name, machineImage.Version); err != nil {
		return append(machineImages, machineImage)
	}
	return machineImages
}
