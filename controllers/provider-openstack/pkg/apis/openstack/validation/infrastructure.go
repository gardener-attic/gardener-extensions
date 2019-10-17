// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package validation

import (
	apisopenstack "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack"

	cidrvalidation "github.com/gardener/gardener/pkg/utils/validation/cidr"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateInfrastructureConfig validates a InfrastructureConfig object.
func ValidateInfrastructureConfig(infra *apisopenstack.InfrastructureConfig, constraints apisopenstack.Constraints, nodesCIDR *string) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(infra.FloatingPoolName) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("floatingPoolName"), "must provide the name of a floating pool"))
	} else if ok, validFloatingPoolNames := validateFloatingPoolNameConstraints(constraints.FloatingPools, infra.FloatingPoolName, ""); !ok {
		allErrs = append(allErrs, field.NotSupported(field.NewPath("floatingPoolName"), infra.FloatingPoolName, validFloatingPoolNames))
	}

	var nodes cidrvalidation.CIDR
	if nodesCIDR != nil {
		nodes = cidrvalidation.NewCIDR(*nodesCIDR, nil)
	}

	networksPath := field.NewPath("networks")
	if len(infra.Networks.Worker) == 0 {
		allErrs = append(allErrs, field.Required(networksPath.Child("worker"), "must specify the network range for the worker network"))
	}

	workerCIDR := cidrvalidation.NewCIDR(infra.Networks.Worker, networksPath.Child("worker"))
	allErrs = append(allErrs, cidrvalidation.ValidateCIDRParse(workerCIDR)...)
	allErrs = append(allErrs, cidrvalidation.ValidateCIDRIsCanonical(networksPath.Child("worker"), infra.Networks.Worker)...)

	if nodes != nil {
		allErrs = append(allErrs, nodes.ValidateSubset(workerCIDR)...)
	}

	if infra.Networks.Router != nil && len(infra.Networks.Router.ID) == 0 {
		allErrs = append(allErrs, field.Invalid(networksPath.Child("router", "id"), infra.Networks.Router.ID, "router id must not be empty when router key is provided"))
	}

	return allErrs
}

// ValidateInfrastructureConfigUpdate validates a InfrastructureConfig object.
func ValidateInfrastructureConfigUpdate(oldConfig, newConfig *apisopenstack.InfrastructureConfig, constraints apisopenstack.Constraints, nodesCIDR *string) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, apivalidation.ValidateImmutableField(newConfig.Networks, oldConfig.Networks, field.NewPath("networks"))...)

	return allErrs
}

func validateFloatingPoolNameConstraints(names []apisopenstack.FloatingPool, name, oldName string) (bool, []string) {
	if name == oldName {
		return true, nil
	}

	validValues := []string{}

	for _, n := range names {
		validValues = append(validValues, n.Name)
		if n.Name == name {
			return true, nil
		}
	}

	return false, validValues
}
