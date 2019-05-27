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

package alicloud

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// Credentials are the credentials to access the Alicloud API.
type Credentials struct {
	AccessKeyID     string
	AccessKeySecret string
}

const (
	// AccessKeyID is the data field in a secret where the access key id is stored at.
	AccessKeyID = "accessKeyID"
	// AccessKeySecret is the data field in a secret where the access key secret is stored at.
	AccessKeySecret = "accessKeySecret"
)

// ReadSecretCredentials reads the Credentials from the given secret.
func ReadSecretCredentials(secret *corev1.Secret) (*Credentials, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret %s/%s has no data section", secret.Namespace, secret.Name)
	}

	accessKeyID, ok := secret.Data[AccessKeyID]
	if !ok {
		return nil, fmt.Errorf("secret %s/%s has no access key id at data.%s", secret.Namespace, secret.Name, AccessKeyID)
	}

	accessKeySecret, ok := secret.Data[AccessKeySecret]
	if !ok {
		return nil, fmt.Errorf("secret %s/%s has no access key secret at data.%s", secret.Namespace, secret.Name, AccessKeySecret)
	}

	return &Credentials{
		AccessKeyID:     string(accessKeyID),
		AccessKeySecret: string(accessKeySecret),
	}, nil
}
