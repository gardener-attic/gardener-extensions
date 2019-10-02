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

package helper

import (
	"fmt"
	"github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere"
	"github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere/validation"
	"github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/common"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func GetCloudProfileConfig(ctx *common.ClientContext, cluster *controller.Cluster) (*vsphere.CloudProfileConfig, error) {
	var cloudProfileConfig *vsphere.CloudProfileConfig
	if cluster != nil && cluster.CloudProfile != nil && cluster.CloudProfile.Spec.ProviderConfig != nil && cluster.CloudProfile.Spec.ProviderConfig.Raw != nil {
		cloudProfileConfig = &vsphere.CloudProfileConfig{}
		if _, _, err := ctx.Decoder().Decode(cluster.CloudProfile.Spec.ProviderConfig.Raw, nil, cloudProfileConfig); err != nil {
			return nil, errors.Wrapf(err, "could not decode providerConfig of cloudProfile for '%s'", cluster.Shoot.Name)
		}
		// TODO validate cloud profile on admission instead
		if errs := validation.ValidateCloudProfileConfig(cloudProfileConfig); len(errs) > 0 {
			return nil, errors.Wrap(errs.ToAggregate(), fmt.Sprintf("validation of providerConfig of cloud profile %q failed", cluster.CloudProfile.Name))
		}
	}
	return cloudProfileConfig, nil
}

func GetControlPlaneConfig(ctx *common.ClientContext, cluster *controller.Cluster) (*vsphere.ControlPlaneConfig, error) {
	cpConfig := &vsphere.ControlPlaneConfig{}
	if cluster.Shoot.Spec.Provider.ControlPlaneConfig != nil {
		if _, _, err := ctx.Decoder().Decode(cluster.Shoot.Spec.Provider.ControlPlaneConfig.Raw, nil, cpConfig); err != nil {
			return nil, errors.Wrapf(err, "could not decode providerConfig of controlplane '%s'", cluster.Shoot.Name)
		}
	}
	return cpConfig, nil
}

func GetInfrastructureStatus(ctx *common.ClientContext, name string, extension *runtime.RawExtension) (*vsphere.InfrastructureStatus, error) {
	infraStatus := &vsphere.InfrastructureStatus{}
	if _, _, err := ctx.Decoder().Decode(extension.Raw, nil, infraStatus); err != nil {
		return nil, errors.Wrapf(err, "could not decode infrastructureProviderStatus of controlplane '%s'", name)
	}
	return infraStatus, nil
}

// InfrastructureConfigFromInfrastructure extracts the InfrastructureConfig from the
// ProviderConfig section of the given Infrastructure.
func GetInfrastructureConfig(ctx *common.ClientContext, cluster *controller.Cluster) (*vsphere.InfrastructureConfig, error) {
	config := &vsphere.InfrastructureConfig{}
	if source := cluster.Shoot.Spec.Provider.InfrastructureConfig; source != nil && source.Raw != nil {
		if _, _, err := ctx.Decoder().Decode(source.Raw, nil, config); err != nil {
			return nil, err
		}
		return config, nil
	}
	return nil, fmt.Errorf("infrastructure config is not set on the infrastructure resource")
}
