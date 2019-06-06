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

package backupentry

import (
	"context"
	"fmt"

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	awsclient "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws/client"
	"github.com/gardener/gardener-extensions/pkg/controller/backupentry"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type actuator struct {
	client client.Client
	logger logr.Logger
}

func newActuator() backupentry.ProviderActuator {
	return &actuator{
		logger: logger,
	}
}

func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

func (a *actuator) GetETCDSecretData(ctx context.Context, be *extensionsv1alpha1.BackupEntry, backupSecretData map[string][]byte) (map[string][]byte, error) {
	backupSecretData[aws.Region] = []byte(be.Spec.Region)
	return backupSecretData, nil
}

func (a *actuator) Delete(ctx context.Context, be *extensionsv1alpha1.BackupEntry) error {
	providerSecret := &corev1.Secret{}
	if err := a.client.Get(ctx, kutil.Key(be.Spec.SecretRef.Namespace, be.Spec.SecretRef.Name), providerSecret); err != nil {
		return err
	}

	awsClient, err := awsclient.NewClient(string(providerSecret.Data[aws.AccessKeyID]), string(providerSecret.Data[aws.SecretAccessKey]), be.Spec.Region)
	if err != nil {
		return err
	}

	return awsClient.DeleteObjectsWithPrefix(ctx, be.Spec.BucketName, fmt.Sprintf("%s/", be.Name))
}
