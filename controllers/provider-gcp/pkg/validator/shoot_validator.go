// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

	"k8s.io/apimachinery/pkg/util/sets"

	gcpvalidation "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp/validation"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener/pkg/apis/garden"
)

// getAllowedRegionZonesFromCloudProfile fetches the set of allowed zones from the Cloud Profile.
func getAllowedRegionZonesFromCloudProfile(shoot *garden.Shoot, cloudProfile *gardencorev1beta1.CloudProfile) sets.String {
	shootRegion := shoot.Spec.Region
	for _, region := range cloudProfile.Spec.Regions {
		if region.Name == shootRegion {
			gcpZones := sets.NewString()
			for _, gcpZone := range region.Zones {
				gcpZones.Insert(gcpZone.Name)
			}
			return gcpZones
		}
	}
	return nil
}

func (v *Shoot) validateShoot(ctx context.Context, shoot *garden.Shoot) error {
	// Network validation
	networkPath := field.NewPath("spec", "networking")

	if errList := gcpvalidation.ValidateNetworking(shoot.Spec.Networking, networkPath); len(errList) != 0 {
		return errList.ToAggregate()
	}

	// Provider validation
	fldPath := field.NewPath("spec", "provider")

	// InfrastructureConfig
	infraConfigFldPath := fldPath.Child("infrastructureConfig")

	if shoot.Spec.Provider.InfrastructureConfig == nil {
		return field.Required(infraConfigFldPath, "InfrastructureConfig must be set for GCP shoots")
	}

	infraConfig, err := decodeInfrastructureConfig(v.decoder, shoot.Spec.Provider.InfrastructureConfig, infraConfigFldPath)
	if err != nil {
		return err
	}

	if errList := gcpvalidation.ValidateInfrastructureConfig(infraConfig, shoot.Spec.Networking.Nodes, shoot.Spec.Networking.Pods, shoot.Spec.Networking.Services); len(errList) != 0 {
		return errList.ToAggregate()
	}

	// Controlplane Config

	cloudProfile := &gardencorev1beta1.CloudProfile{}
	if err := v.client.Get(ctx, kutil.Key(shoot.Spec.CloudProfileName), cloudProfile); err != nil {
		return err
	}

	var (
		allowedZones = getAllowedRegionZonesFromCloudProfile(shoot, cloudProfile)
		workersPath  = fldPath.Child("workers")
	)

	controlPlaneConfigPath := fldPath.Child("controlPlaneConfig")
	if shoot.Spec.Provider.ControlPlaneConfig == nil {
		return field.Required(controlPlaneConfigPath, "ControlPlaneConfig must be set for GCP shoots")
	}

	controlPlaneConfig, err := decodeControlPlaneConfig(v.decoder, shoot.Spec.Provider.ControlPlaneConfig, controlPlaneConfigPath)
	if err != nil {
		return err
	}

	if errList := gcpvalidation.ValidateControlPlaneConfig(controlPlaneConfig, allowedZones); len(errList) != 0 {
		return errList.ToAggregate()
	}

	if errList := gcpvalidation.ValidateWorkers(shoot.Spec.Provider.Workers, allowedZones, workersPath); len(errList) != 0 {
		return errList.ToAggregate()
	}

	return nil
}

func (v *Shoot) validateShootUpdate(ctx context.Context, oldShoot, shoot *garden.Shoot) error {
	fldPath := field.NewPath("spec", "provider")

	// InfrastructureConfig
	if shoot.Spec.Provider.InfrastructureConfig == nil {
		return field.Required(fldPath.Child("infrastructureConfig"), "InfrastructureConfig must be set for GCP shoots")
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
		if errList := gcpvalidation.ValidateInfrastructureConfigUpdate(oldInfraConfig, infraConfig); len(errList) != 0 {
			return errList.ToAggregate()
		}
	}

	// ControlPlaneConfig
	if shoot.Spec.Provider.ControlPlaneConfig == nil {
		return field.Required(fldPath.Child("controlPlaneConfig"), "controlPlaneConfig must be set for GCP shoots")
	}

	if oldShoot.Spec.Provider.ControlPlaneConfig == nil {
		return field.InternalError(fldPath.Child("controlPlaneConfig"), errors.New("controlPlaneConfig is not available on old shoot"))
	}

	controlPlaneConfig, err := decodeControlPlaneConfig(v.decoder, shoot.Spec.Provider.InfrastructureConfig, fldPath)
	if err != nil {
		return err
	}

	oldControlPlaneConfig, err := decodeControlPlaneConfig(v.decoder, oldShoot.Spec.Provider.ControlPlaneConfig, fldPath)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(oldControlPlaneConfig, controlPlaneConfig) {
		if errList := gcpvalidation.ValidateControlPlaneConfigUpdate(oldControlPlaneConfig, controlPlaneConfig); len(errList) != 0 {
			return errList.ToAggregate()
		}
	}

	return v.validateShoot(ctx, shoot)
}
