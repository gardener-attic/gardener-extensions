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

package infrastructure

import (
	"context"
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/imagevector"

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	glogger "github.com/gardener/gardener/pkg/logger"
	"github.com/gardener/gardener/pkg/operation/terraformer"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type actuator struct {
	logger logr.Logger

	restConfig *rest.Config

	client  client.Client
	scheme  *runtime.Scheme
	decoder runtime.Decoder
}

// NewActuator creates a new Actuator that updates the status of the handled Infrastructure resources.
func NewActuator() infrastructure.Actuator {
	return &actuator{
		logger: log.Log.WithName("infrastructure-actuator"),
	}
}

func (a *actuator) InjectScheme(scheme *runtime.Scheme) error {
	a.scheme = scheme
	a.decoder = serializer.NewCodecFactory(a.scheme).UniversalDecoder()
	return nil
}

func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

func (a *actuator) InjectConfig(config *rest.Config) error {
	a.restConfig = config
	return nil
}

func (a *actuator) Reconcile(ctx context.Context, config *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	return a.reconcile(ctx, config, cluster)
}

func (a *actuator) Delete(ctx context.Context, config *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	return a.delete(ctx, config, cluster)
}

// Helper functions

func (a *actuator) newTerraformer(purpose, namespace, name string) (*terraformer.Terraformer, error) {
	return terraformer.NewForConfig(glogger.NewLogger("info"), a.restConfig, purpose, namespace, name, imagevector.TerraformerImage())
}

func generateTerraformInfraVariablesEnvironment(secret *corev1.Secret) map[string]string {
	return terraformer.GenerateVariablesEnvironment(secret, map[string]string{
		"ACCESS_KEY_ID":     aws.AccessKeyID,
		"SECRET_ACCESS_KEY": aws.SecretAccessKey,
	})
}
