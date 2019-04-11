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
	"context"
	"fmt"

	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// SubscriptionIDKey is the key for the subscription ID
	SubscriptionIDKey = "subscriptionID"
	// TenantIDKey is the key for the tenant id
	TenantIDKey = "tenantID"
	// ClientIDKey is the key for the client id
	ClientIDKey = "clientID"
	// ClientSecretKey is the key for the client secret
	ClientSecretKey = "clientSecret"
)

// ClientAuth represents a Azure Client Auth credentials.
type ClientAuth struct {
	// SubscriptionID is the azure subscription ID.
	SubscriptionID string
	// TenantID is the azure tenant id.
	TenantID string
	// ClientID is the azure client id
	ClientID string
	// ClientSecret is the client secret
	ClientSecret string
}

// GetClientAuthData retrieves the specified service account.
func GetClientAuthData(ctx context.Context, c client.Client, namespace, name string) (*ClientAuth, error) {
	secret := &corev1.Secret{}
	if err := c.Get(ctx, kutil.Key(namespace, name), secret); err != nil {
		return nil, err
	}

	return ReadClientAuthDataFromSecret(secret)
}

// ReadClientAuthDataFromSecret reads the client auth details from the given secret.
func ReadClientAuthDataFromSecret(secret *corev1.Secret) (*ClientAuth, error) {
	subscriptionID, ok := secret.Data[SubscriptionIDKey]
	if !ok {
		return nil, fmt.Errorf("secret %s/%s doesn't have a subscription ID", secret.Namespace, secret.Name)
	}

	clientID, ok := secret.Data[ClientIDKey]
	if !ok {
		return nil, fmt.Errorf("secret %s/%s doesn't have a client ID", secret.Namespace, secret.Name)
	}

	tenantID, ok := secret.Data[TenantIDKey]
	if !ok {
		return nil, fmt.Errorf("secret %s/%s doesn't have a tenant ID", secret.Namespace, secret.Name)
	}

	clientSecret, ok := secret.Data[ClientSecretKey]
	if !ok {
		return nil, fmt.Errorf("secret %s/%s doesn't have a Client Secret", secret.Namespace, secret.Name)
	}

	return &ClientAuth{
		SubscriptionID: string(subscriptionID),
		ClientID:       string(clientID),
		TenantID:       string(tenantID),
		ClientSecret:   string(clientSecret),
	}, nil
}
