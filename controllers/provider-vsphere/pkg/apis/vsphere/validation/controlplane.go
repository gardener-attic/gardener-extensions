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

package validation

import (
	apisvsphere "github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateControlPlaneConfig validates a ControlPlaneConfig object.
func ValidateControlPlaneConfig(controlPlaneConfig *apisvsphere.ControlPlaneConfig, region string, regions []gardencorev1beta1.Region, constraints apisvsphere.Constraints) field.ErrorList {
	allErrs := field.ErrorList{}

	knownClassNames := map[string]bool{}
	for _, cls := range constraints.LoadBalancerConfig.Classes {
		knownClassNames[cls.Name] = true
	}
	for _, lbclass := range controlPlaneConfig.LoadBalancerClasses {
		if lbclass.Name == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("loadBalancerClasses", "name"), "name of load balancer class must be set"))
		} else if _, ok := knownClassNames[lbclass.Name]; !ok {
			allErrs = append(allErrs, field.Invalid(field.NewPath("loadBalancerClasses", "name"), lbclass.Name, "must be defined in in cloud profile constraints"))
		}
	}

	return allErrs
}

// ValidateControlPlaneConfigUpdate validates a ControlPlaneConfig object.
func ValidateControlPlaneConfigUpdate(oldConfig, newConfig *apisvsphere.ControlPlaneConfig, region string, regions []gardencorev1beta1.Region) field.ErrorList {
	allErrs := field.ErrorList{}
	return allErrs
}
