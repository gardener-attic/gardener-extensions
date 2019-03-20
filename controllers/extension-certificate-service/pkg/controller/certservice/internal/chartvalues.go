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
	"strings"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config"
)

// CreateDNSProviderValue creates values for the passed DNSProviderConfig if the provider manages the passed shootDomain.
func CreateDNSProviderValue(config config.DNSProviderConfig, shootDomain string) map[string]string {
	var dnsConfig map[string]string
	for _, domain := range config.DomainNames() {
		if shootDomain == domain || strings.HasSuffix(shootDomain, "."+domain) {
			dnsConfig = make(map[string]string)
			dnsConfig["domain"] = shootDomain
			dnsConfig["provider"] = config.ProviderName()
			break
		}
	}
	return dnsConfig
}
