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

package util

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gardener/gardener/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ContextFromStopChannel creates a new context from a given stop channel.
func ContextFromStopChannel(stopCh <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		<-stopCh
	}()

	return ctx
}

// ComputeChecksum computes a SHA256 checksum for the give map.
func ComputeChecksum(data interface{}) string {
	jsonString, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return utils.ComputeSHA256Hex(jsonString)
}

// GetKubeconfigFromSecret gets the Kubeconfig from the passed secret.
func GetKubeconfigFromSecret(secret *corev1.Secret) (*restclient.Config, error) {
	var (
		key            = "kubeconfig"
		kubeconfig, ok = secret.Data[key]
	)
	if !ok {
		return nil, fmt.Errorf("Key %s not available in map", key)
	}
	return clientcmd.RESTConfigFromKubeConfig(kubeconfig)
}
