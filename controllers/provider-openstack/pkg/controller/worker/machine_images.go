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

package worker

import (
	"context"
	"fmt"

	confighelper "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/config/helper"
	apisopenstack "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack"
	apisopenstackhelper "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack/helper"
	openstackv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack/v1alpha1"
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
		workerStatus = &apisopenstack.WorkerStatus{
			TypeMeta: metav1.TypeMeta{
				APIVersion: apisopenstack.SchemeGroupVersion.String(),
				Kind:       "WorkerStatus",
			},
			MachineImages: w.machineImages,
		}

		workerStatusV1alpha1 = &openstackv1alpha1.WorkerStatus{
			TypeMeta: metav1.TypeMeta{
				APIVersion: openstackv1alpha1.SchemeGroupVersion.String(),
				Kind:       "WorkerStatus",
			},
		}
	)

	if err := w.scheme.Convert(workerStatus, workerStatusV1alpha1, nil); err != nil {
		return nil, err
	}

	return workerStatusV1alpha1, nil
}

func (w *workerDelegate) findMachineImage(name, version, cloudProfile string) (string, error) {
	ami, err := confighelper.FindImageForCloudProfile(w.machineImageToCloudProfilesMapping, name, version, cloudProfile)
	if err == nil {
		return ami, nil
	}

	// Try to look up machine image in worker provider status as it was not found in componentconfig.
	if providerStatus := w.worker.Status.ProviderStatus; providerStatus != nil {
		workerStatus := &apisopenstack.WorkerStatus{}
		if _, _, err := w.decoder.Decode(providerStatus.Raw, nil, workerStatus); err != nil {
			return "", errors.Wrapf(err, "could not decode worker status of worker '%s'", util.ObjectName(w.worker))
		}

		machineImage, err := apisopenstackhelper.FindMachineImage(workerStatus.MachineImages, name, version)
		if err != nil {
			return "", errorMachineImageNotFound(name, version, cloudProfile)
		}

		return machineImage.Image, nil
	}

	return "", errorMachineImageNotFound(name, version, cloudProfile)
}

func errorMachineImageNotFound(name, version, cloudProfile string) error {
	return fmt.Errorf("could not find machine image for %s/%s/%s neither in componentconfig nor in worker status", name, version, cloudProfile)
}

func appendMachineImage(machineImages []apisopenstack.MachineImage, machineImage apisopenstack.MachineImage) []apisopenstack.MachineImage {
	if _, err := apisopenstackhelper.FindMachineImage(machineImages, machineImage.Name, machineImage.Version); err != nil {
		return append(machineImages, machineImage)
	}
	return machineImages
}
