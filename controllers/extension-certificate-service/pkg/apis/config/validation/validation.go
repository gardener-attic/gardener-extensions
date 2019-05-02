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

package validation

import (
	"net/url"
	"time"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config"

	"github.com/gardener/gardener/pkg/utils"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateConfiguration validates the passed configuration instance.
func ValidateConfiguration(config *config.Configuration) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateConfigurationSpec(&config.Spec, field.NewPath("spec"))...)

	return allErrs
}

func validateConfigurationSpec(spec *config.ConfigurationSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if spec.LifecycleSync.Duration <= time.Duration(0) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("lifecycleSync"), spec.LifecycleSync, "must be greater than 0"))
	}

	if spec.ServiceSync.Duration <= time.Duration(0) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("serviceSync"), spec.LifecycleSync, "must be greater than 0"))
	}

	if spec.IssuerName == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("issuerName"), "field is required"))
	}

	if spec.NamespaceRef == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("namespaceRef"), "field is required"))
	}

	if spec.ResourceNamespace == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("resourceNamespace"), "field is required"))
	}

	allErrs = append(allErrs, validateACME(&spec.ACME, fldPath.Child("acme"))...)
	allErrs = append(allErrs, validateDNSProviders(&spec.Providers, fldPath.Child("providers"))...)

	return allErrs
}

func validateACME(acme *config.ACME, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if _, err := url.ParseRequestURI(acme.Server); err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("server"), acme.Server, err.Error()))
	}

	if !utils.TestEmail(acme.Email) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("email"), acme.Email, "must be a valid mail address"))
	}

	return allErrs
}

func validateDNSProviders(providers *config.DNSProviders, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for i, route53 := range providers.Route53 {
		provider := &route53
		allErrs = append(allErrs, validateRoute53Provider(provider, fldPath.Child("route53").Index(i))...)
	}

	for i, cloudDNS := range providers.CloudDNS {
		provider := &cloudDNS
		allErrs = append(allErrs, validateCloudDNSProvider(provider, fldPath.Child("clouddns").Index(i))...)
	}

	return allErrs
}

func validateRoute53Provider(route53 *config.Route53, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if route53.Name == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), "field is required"))
	}

	if route53.Region == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("region"), "field is required"))
	}

	if route53.AccessKeyID == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("accessKeyID"), "field is required"))
	}

	if route53.SecretAccessKey == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("secretAccessKey"), "field is required"))
	}

	return allErrs
}

func validateCloudDNSProvider(cloudDNS *config.CloudDNS, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if cloudDNS.Name == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), "field is required"))
	}

	if cloudDNS.Project == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("project"), "field is required"))
	}

	if cloudDNS.ServiceAccount == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("serviceAccount"), "field is required"))
	}

	return allErrs
}
