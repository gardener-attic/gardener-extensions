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
	"fmt"

	"github.com/gardener/gardener/pkg/operation/common"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/gardener/gardener/pkg/utils/secrets"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const caChecksumAnnotation = "checksum/ca"

// GetOrCreateShootKubeconfig gets or creates a Kubeconfig for a Shoot cluster which has a running control plane in the given `namespace`.
// If the CA of an existing Kubeconfig has changed, it creates a new Kubeconfig.
// Newly generated Kubeconfigs are applied with the given `client` to the given `namespace`.
func GetOrCreateShootKubeconfig(ctx context.Context, client client.Client, certificateConfig secrets.CertificateSecretConfig, namespace string) (*corev1.Secret, error) {
	caSecret, ca, err := secrets.LoadCAFromSecret(client, namespace, gardencorev1alpha1.SecretNameCACluster)
	if err != nil {
		return nil, fmt.Errorf("error fetching CA secret %s/%s: %v", namespace, gardencorev1alpha1.SecretNameCACluster, err)
	}

	var (
		secret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: make(map[string]string),
			},
		}
		key = types.NamespacedName{
			Name:      certificateConfig.Name,
			Namespace: namespace,
		}
	)
	if err := client.Get(ctx, key, &secret); err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("error preparing kubeconfig: %v", err)
	}

	var (
		computedChecksum   = ComputeSecretCheckSum(caSecret.Data)
		storedChecksum, ok = secret.Annotations[caChecksumAnnotation]
	)
	if ok && computedChecksum == storedChecksum {
		return &secret, nil
	}

	certificateConfig.SigningCA = ca
	certificateConfig.CertType = secrets.ClientCert

	config := secrets.ControlPlaneSecretConfig{
		CertificateSecretConfig: &certificateConfig,

		KubeConfigRequest: &secrets.KubeConfigRequest{
			ClusterName:  namespace,
			APIServerURL: kubeAPIServerServiceDNS(namespace),
		},
	}

	controlPlane, err := config.GenerateControlPlane()
	if err != nil {
		return nil, fmt.Errorf("error creating kubeconfig: %v", err)
	}

	return &secret, kubernetes.CreateOrUpdate(ctx, client, &secret, func() error {
		secret.ObjectMeta.Name = certificateConfig.Name
		secret.ObjectMeta.Namespace = namespace
		secret.Data = controlPlane.SecretData()
		secret.Annotations[caChecksumAnnotation] = computedChecksum
		return nil
	})
}

// KubeAPIServerServiceDNS returns a domain name which can be used to contact
// the Kube-Apiserver deployment of a Shoot within the Seed cluster.
// e.g. kube-apiserver.shoot--project--prod.svc.cluster.local.
func kubeAPIServerServiceDNS(namespace string) string {
	return fmt.Sprintf("%s.%s.%s", common.KubeAPIServerDeploymentName, namespace, "svc.cluster.local.")
}

// GetReplicaCount returns the given replica count base on the hibernation status of the shoot.
func GetReplicaCount(shoot *gardenv1beta1.Shoot, count int) int {
	if shoot.Spec.Hibernation != nil && shoot.Spec.Hibernation.Enabled {
		return 0
	}
	return count
}
