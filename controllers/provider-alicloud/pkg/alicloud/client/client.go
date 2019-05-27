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

package client

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"

	alicloudvpc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
)

const (
	// AccessKeyIDField is the data field in a secret where the access key id is stored at.
	AccessKeyIDField = "accessKeyID"
	// AccessKeySecretField is the data field in a secret where the access key secret is stored at.
	AccessKeySecretField = "accessKeySecret"
)

// ReadSecretCredentials reads the Credentials from the given secret.
func ReadSecretCredentials(secret *corev1.Secret) (*Credentials, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret %s/%s has no data section", secret.Namespace, secret.Name)
	}

	accessKeyID, ok := secret.Data[AccessKeyIDField]
	if !ok {
		return nil, fmt.Errorf("secret %s/%s has no access key id at data.%s", secret.Namespace, secret.Name, AccessKeyIDField)
	}

	accessKeySecret, ok := secret.Data[AccessKeySecretField]
	if !ok {
		return nil, fmt.Errorf("secret %s/%s has no access key secret at data.%s", secret.Namespace, secret.Name, AccessKeySecretField)
	}

	return &Credentials{
		AccessKeyID:     string(accessKeyID),
		AccessKeySecret: string(accessKeySecret),
	}, nil
}

// FactoryFunc is a function that implements the Factory interface. Used for consuming the
// `alicloudvpc.NewClientWithAccessKey` function.
type FactoryFunc func(region, accessKeyID, accessKeySecret string) (*alicloudvpc.Client, error)

// NewVPC implements Factory.
func (f FactoryFunc) NewVPC(region, accessKeyID, accessKeySecret string) (VPC, error) {
	return f(region, accessKeyID, accessKeySecret)
}

// DefaultFactory instantiates a default Factory.
func DefaultFactory() Factory {
	return FactoryFunc(alicloudvpc.NewClientWithAccessKey)
}
