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
	apisgcp "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp"

	cidrvalidation "github.com/gardener/gardener/pkg/utils/validation/cidr"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateInfrastructureConfig validates a InfrastructureConfig object.
func ValidateInfrastructureConfig(infra *apisgcp.InfrastructureConfig, nodesCIDR, podsCIDR, servicesCIDR *string) field.ErrorList {
	allErrs := field.ErrorList{}

	var (
		nodes    cidrvalidation.CIDR
		pods     cidrvalidation.CIDR
		services cidrvalidation.CIDR
	)

	if nodesCIDR != nil {
		nodes = cidrvalidation.NewCIDR(*nodesCIDR, nil)
	}
	if podsCIDR != nil {
		pods = cidrvalidation.NewCIDR(*podsCIDR, nil)
	}
	if servicesCIDR != nil {
		services = cidrvalidation.NewCIDR(*servicesCIDR, nil)
	}

	networksPath := field.NewPath("networks")
	if len(infra.Networks.Worker) == 0 && len(infra.Networks.Workers) == 0 {
		allErrs = append(allErrs, field.Required(networksPath.Child("workers"), "must specify the network range for the worker network"))
	}

	var workerCIDR cidrvalidation.CIDR
	if infra.Networks.Worker != "" {
		workerCIDR = cidrvalidation.NewCIDR(infra.Networks.Worker, networksPath.Child("worker"))
		allErrs = append(allErrs, cidrvalidation.ValidateCIDRParse(workerCIDR)...)
		allErrs = append(allErrs, cidrvalidation.ValidateCIDRIsCanonical(networksPath.Child("worker"), infra.Networks.Worker)...)
	}
	if infra.Networks.Workers != "" {
		workerCIDR = cidrvalidation.NewCIDR(infra.Networks.Workers, networksPath.Child("workers"))
		allErrs = append(allErrs, cidrvalidation.ValidateCIDRParse(workerCIDR)...)
		allErrs = append(allErrs, cidrvalidation.ValidateCIDRIsCanonical(networksPath.Child("workers"), infra.Networks.Workers)...)
	}

	if infra.Networks.Internal != nil {
		internalCIDR := cidrvalidation.NewCIDR(*infra.Networks.Internal, networksPath.Child("internal"))
		allErrs = append(allErrs, cidrvalidation.ValidateCIDRParse(internalCIDR)...)
		allErrs = append(allErrs, cidrvalidation.ValidateCIDRIsCanonical(networksPath.Child("internal"), *infra.Networks.Internal)...)
		allErrs = append(allErrs, cidrvalidation.ValidateCIDROverlap([]cidrvalidation.CIDR{pods, services}, []cidrvalidation.CIDR{internalCIDR}, false)...)
		allErrs = append(allErrs, cidrvalidation.ValidateCIDROverlap([]cidrvalidation.CIDR{nodes}, []cidrvalidation.CIDR{internalCIDR}, false)...)
		allErrs = append(allErrs, cidrvalidation.ValidateCIDROverlap([]cidrvalidation.CIDR{workerCIDR}, []cidrvalidation.CIDR{internalCIDR}, false)...)
	}

	if nodes != nil {
		allErrs = append(allErrs, nodes.ValidateSubset(workerCIDR)...)
	}

	if infra.Networks.VPC != nil && len(infra.Networks.VPC.Name) == 0 {
		allErrs = append(allErrs, field.Invalid(networksPath.Child("vpc", "name"), infra.Networks.VPC.Name, "vpc name must not be empty when vpc key is provided"))
	}

	if infra.Networks.VPC != nil && len(infra.Networks.VPC.Name) == 0 && infra.Networks.VPC.CloudRouter != nil {
		allErrs = append(allErrs, field.Invalid(networksPath.Child("vpc", "cloudRouter"), infra.Networks.VPC.CloudRouter, "cloud router can not be configured when the VPC name is not specified"))
	}

	return allErrs
}

// ValidateInfrastructureConfigUpdate validates a InfrastructureConfig object.
func ValidateInfrastructureConfigUpdate(oldConfig, newConfig *apisgcp.InfrastructureConfig, nodesCIDR, podsCIDR, servicesCIDR *string) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, apivalidation.ValidateImmutableField(newConfig.Networks, oldConfig.Networks, field.NewPath("networks"))...)

	return allErrs
}
