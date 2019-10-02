/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package internal

import (
	"context"
	"fmt"

	"github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/vsphere"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Credentials contains the necessary vSphere credential information.
type Credentials struct {
	VsphereUsername string
	VspherePassword string

	NSXTUsername string
	NSXTPassword string
}

// GetCredentials computes for a given context and infrastructure the corresponding credentials object.
func GetCredentials(ctx context.Context, c client.Client, secretRef corev1.SecretReference) (*Credentials, error) {
	secret, err := extensionscontroller.GetSecretByReference(ctx, c, &secretRef)
	if err != nil {
		return nil, err
	}
	return ExtractCredentials(secret)
}

// ExtractCredentials generates a credentials object for a given provider secret.
func ExtractCredentials(secret *corev1.Secret) (*Credentials, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret does not contain any data")
	}
	username, ok := secret.Data[vsphere.Username]
	if !ok {
		return nil, fmt.Errorf("missing %q field in secret", vsphere.Username)
	}

	password, ok := secret.Data[vsphere.Password]
	if !ok {
		return nil, fmt.Errorf("missing %q field in secret", vsphere.Password)
	}

	nsxtUsername, ok := secret.Data[vsphere.NSXTUsername]
	if !ok {
		return nil, fmt.Errorf("missing %q field in secret", vsphere.NSXTUsername)
	}

	nsxtPassword, ok := secret.Data[vsphere.NSXTPassword]
	if !ok {
		return nil, fmt.Errorf("missing %q field in secret", vsphere.NSXTPassword)
	}

	return &Credentials{
		VsphereUsername: string(username),
		VspherePassword: string(password),
		NSXTUsername:    string(nsxtUsername),
		NSXTPassword:    string(nsxtPassword),
	}, nil
}
