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

package internal

import (
	"fmt"

	apisconfig "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config"
	"k8s.io/apimachinery/pkg/types"
)

// CreateCertServiceValues creates chart values for the certificate service.
func CreateCertServiceValues(certmanagementConfig apisconfig.ConfigurationSpec, namespace string, uid types.UID) (map[string]interface{}, error) {
	var (
		acmeConfig     = certmanagementConfig.ACME
		route53Config  = certmanagementConfig.Providers.Route53
		clouddnsConfig = certmanagementConfig.Providers.CloudDNS
	)

	var dnsProviders []apisconfig.DNSProviderConfig
	for _, route53provider := range route53Config {
		it := route53provider
		dnsProviders = append(dnsProviders, &it)
	}
	for _, cloudDNSProvider := range clouddnsConfig {
		it := cloudDNSProvider
		dnsProviders = append(dnsProviders, &it)
	}

	var (
		letsEncryptSecretName = "lets-encrypt"
		acmePrivateKey        string
	)

	if acmeConfig.PrivateKey != nil {
		acmePrivateKey = *acmeConfig.PrivateKey
	}

	providers, err := CreateDNSProviderValues(dnsProviders)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"clusterissuer": map[string]interface{}{
			"name": certmanagementConfig.IssuerName,
			"acme": map[string]interface{}{
				"email":  acmeConfig.Email,
				"server": acmeConfig.Server,
				"letsEncrypt": map[string]interface{}{
					"name": letsEncryptSecretName,
					"key":  acmePrivateKey,
				},
				"dns01": map[string]interface{}{
					"providers": providers,
				},
			},
		},
		"namespaceRef": map[string]interface{}{
			"namespace": namespace,
			"uid":       uid,
		},
	}, nil
}

// CreateDNSProviderValues creates chart values for the DNS resolvers.
func CreateDNSProviderValues(configs []apisconfig.DNSProviderConfig) ([]map[string]interface{}, error) {
	var providers []map[string]interface{}
	for _, config := range configs {
		name := config.ProviderName()
		switch config.DNSProvider() {
		case apisconfig.Route53Provider:
			route53config, ok := config.(*apisconfig.Route53)
			if !ok {
				return nil, fmt.Errorf("Failed to cast to Route53Config object for DNSProviderConfig  %+v", config)
			}

			providers = append(providers, map[string]interface{}{
				"name":        name,
				"type":        apisconfig.Route53Provider,
				"region":      route53config.Region,
				"accessKeyID": route53config.AccessKeyID,
				"accessKey":   route53config.AccessKey(),
			})
		case apisconfig.CloudDNSProvider:
			cloudDNSConfig, ok := config.(*apisconfig.CloudDNS)
			if !ok {
				return nil, fmt.Errorf("Failed to cast to CloudDNSConfig object for DNSProviderConfig  %+v", config)
			}

			providers = append(providers, map[string]interface{}{
				"name":      name,
				"type":      apisconfig.CloudDNSProvider,
				"project":   cloudDNSConfig.Project,
				"accessKey": cloudDNSConfig.AccessKey(),
			})
		default:
		}
	}
	return providers, nil
}
