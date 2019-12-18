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

package validator

import (
	"context"
	"errors"
	"reflect"

	awsvalidation "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws/validation"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func (v *Shoot) validateShoot(ctx context.Context, shoot *gardencorev1beta1.Shoot) error {
	fldPath := field.NewPath("spec", "provider")

	// InfrastructureConfig
	infraConfigFldPath := fldPath.Child("infrastructureConfig")
	if shoot.Spec.Provider.InfrastructureConfig == nil {
		return field.Required(infraConfigFldPath, "InfrastructureConfig must be set for AWS shoots")
	}

	infraConfig, err := decodeInfrastructureConfig(v.decoder, shoot.Spec.Provider.InfrastructureConfig, infraConfigFldPath)
	if err != nil {
		return err
	}

	errList := awsvalidation.ValidateInfrastructureConfig(infraConfig, &shoot.Spec.Networking.Nodes, shoot.Spec.Networking.Pods, shoot.Spec.Networking.Services)
	if len(errList) != 0 {
		return errList.ToAggregate()
	}

	cloudProfile := &gardencorev1beta1.CloudProfile{}
	if err := v.client.Get(ctx, kutil.Key(shoot.Spec.CloudProfileName), cloudProfile); err != nil {
		return err
	}

	errList = awsvalidation.ValidateInfrastructureConfigAgainstCloudProfile(infraConfig, shoot, cloudProfile, infraConfigFldPath)
	if len(errList) != 0 {
		return errList.ToAggregate()
	}

	// ControlPlaneConfig
	if shoot.Spec.Provider.ControlPlaneConfig != nil {
		_, err = decodeControlPlaneConfig(v.decoder, shoot.Spec.Provider.ControlPlaneConfig, fldPath.Child("controlPlaneConfig"))
		if err != nil {
			return err
		}
	}

	// WorkerConfig
	fldPath = fldPath.Child("workers")
	for i, worker := range shoot.Spec.Provider.Workers {
		if worker.ProviderConfig != nil {
			workerConfig, err := decodeWorkerConfig(v.decoder, worker.ProviderConfig, fldPath.Index(i).Child("providerConfig"))
			if err != nil {
				return err
			}

			var volumeType *string
			if worker.Volume != nil {
				volumeType = worker.Volume.Type
			}
			errList := awsvalidation.ValidateWorkerConfig(workerConfig, volumeType)
			if len(errList) != 0 {
				return errList.ToAggregate()
			}
		}
	}

	// Shoot workers
	errList = awsvalidation.ValidateWorkers(shoot.Spec.Provider.Workers, infraConfig.Networks.Zones, fldPath)
	if len(errList) != 0 {
		return errList.ToAggregate()
	}
	return nil
}

func (v *Shoot) validateShootUpdate(ctx context.Context, oldShoot, shoot *gardencorev1beta1.Shoot) error {
	fldPath := field.NewPath("spec", "provider")

	// InfrastructureConfig update
	if shoot.Spec.Provider.InfrastructureConfig == nil {
		return field.Required(fldPath.Child("infrastructureConfig"), "InfrastructureConfig must be set for AWS shoots")
	}

	infraConfig, err := decodeInfrastructureConfig(v.decoder, shoot.Spec.Provider.InfrastructureConfig, fldPath)
	if err != nil {
		return err
	}

	if oldShoot.Spec.Provider.InfrastructureConfig == nil {
		return field.InternalError(fldPath.Child("infrastructureConfig"), errors.New("InfrastructureConfig is not available on old shoot"))
	}

	oldInfraConfig, err := decodeInfrastructureConfig(v.decoder, oldShoot.Spec.Provider.InfrastructureConfig, fldPath)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(oldInfraConfig, infraConfig) {
		errList := awsvalidation.ValidateInfrastructureConfigUpdate(oldInfraConfig, infraConfig)
		if len(errList) != 0 {
			return errList.ToAggregate()
		}
	}

	return v.validateShoot(ctx, shoot)
}
