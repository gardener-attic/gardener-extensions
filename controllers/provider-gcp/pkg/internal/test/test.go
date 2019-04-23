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

package test

import (
	"context"
	"fmt"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/gcp"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MkServiceAccountData creates service account data with the given project ID.
func MkServiceAccountData(projectID string) []byte {
	return []byte(fmt.Sprintf(`{"project_id": "%s"}`, projectID))
}

// MkServiceAccountSecret creates corev1.Secret representing a service account.
func MkServiceAccountSecret(namespace, name string, data []byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: map[string][]byte{
			gcp.ServiceAccountJSONField: data,
		},
	}
}

// ExpectGetServiceAccountSecret sets up an expectation for a get service account call.
func ExpectGetServiceAccountSecret(ctx context.Context, c *mockclient.MockClient, secret *corev1.Secret) *gomock.Call {
	return c.EXPECT().Get(ctx, kutil.Key(secret.Namespace, secret.Name), gomock.AssignableToTypeOf(&corev1.Secret{})).SetArg(2, *secret)
}
