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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertificateServiceConfiguration contains information about the certificate service configuration.
type CertificateServiceConfiguration struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CertificateServiceConfigurationSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertificateServiceConfigurationList is a list of CertificateServiceConfiguration resources.
type CertificateServiceConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of CertificateServiceConfiguration.
	Items []CertificateServiceConfiguration `json:"items"`
}

// CertificateServiceConfigurationSpec contains information about the certificate service configuration.
type CertificateServiceConfigurationSpec struct {
	IssuerName string       `json:"issuerName"`
	ACME       ACME         `json:"acme"`
	Providers  DNSProviders `json:"providers"`
}

// ACME holds information about the ACME issuer used for the certificate service.
type ACME struct {
	Email      string  `json:"email"`
	Server     string  `json:"server"`
	PrivateKey *string `json:"privateKey,omitempty"`
}

// DNSProviders hold information about information about DNS providers used for ACME DNS01 challenges.
type DNSProviders struct {
	Route53  []Route53  `json:"route53,omitempty"`
	CloudDNS []CloudDNS `json:"clouddns,omitempty"`
}

// Route53 is a DNS provider used for ACME DNS01 challenges.
type Route53 struct {
	Domains         []string `json:"domains"`
	Name            string   `json:"name"`
	Region          string   `json:"region"`
	AccessKeyID     string   `json:"accessKeyID"`
	SecretAccessKey string   `json:"secretAccessKey"`
}

// CloudDNS is a DNS provider used for ACME DNS01 challenges.
type CloudDNS struct {
	Domains        []string `json:"domains"`
	Name           string   `json:"name"`
	Project        string   `json:"project"`
	ServiceAccount string   `json:"servieAccount"`
}

// DNSProviderConfig is an interface that will implemented by cloud provider structs
type DNSProviderConfig interface {
	DNSProvider() DNSProvider
	AccessKey() string
	ProviderName() string
	DomainNames() []string
}

// DNSProvider string type
type DNSProvider string

const (
	// Route53Provider is a constant string for aws-route53.
	Route53Provider DNSProvider = "aws-route53"
	// CloudDNSProvider is a constant string for google-clouddns.
	CloudDNSProvider DNSProvider = "google-clouddns"
)

// DNSProvider returns the provider type  in-use.
func (r *Route53) DNSProvider() DNSProvider {
	return Route53Provider
}

// AccessKey returns the route53 SecretAccessKey in case route53 provider is used.
func (r *Route53) AccessKey() string {
	return r.SecretAccessKey
}

// ProviderName returns the route53 provider name.
func (r *Route53) ProviderName() string {
	return r.Name
}

// DomainNames returns the domains this provider manages.
func (r *Route53) DomainNames() []string {
	return r.Domains
}

// DNSProvider returns the provider type in-use.
func (c *CloudDNS) DNSProvider() DNSProvider {
	return CloudDNSProvider
}

// AccessKey returns the CloudDNS ServiceAccount in case Google CloudDNS provider is used.
func (c *CloudDNS) AccessKey() string {
	return c.ServiceAccount
}

// ProviderName returns the CloudDNS provider name.
func (c *CloudDNS) ProviderName() string {
	return c.Name
}

// DomainNames returns the domains this provider manages.
func (c *CloudDNS) DomainNames() []string {
	return c.Domains
}
