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
	"fmt"
	"path/filepath"
	"time"

	awsapi "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/awsclient"

	gardenerLogger "github.com/gardener/gardener/pkg/logger"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/tools/record"

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"
	"github.com/gardener/gardener/pkg/utils/flow"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/operation/terraformer"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	loggerName    = "infrastructure-actuator"
	accessKeysMap = map[string]string{
		"ACCESS_KEY_ID":     aws.AccessKeyID,
		"SECRET_ACCESS_KEY": aws.SecretAccessKey,
	}
)

type actuator struct {
	config           *rest.Config
	recorder         record.EventRecorder
	client           client.Client
	kubernetes       kubernetes.Interface
	scheme           *runtime.Scheme
	logger           logr.Logger
	decoder          runtime.Decoder
	serverVersion    *version.Info
	terraformerImage string
}

// NewActuator creates a new Actuator that updates the status of the handled Infrastructure resources.
func NewActuator(eventRecorder record.EventRecorder, terraformerImage string) infrastructure.Actuator {
	return &actuator{
		logger:           log.Log.WithName(loggerName),
		recorder:         eventRecorder,
		terraformerImage: terraformerImage,
	}
}

func (a *actuator) InjectScheme(scheme *runtime.Scheme) error {
	a.scheme = scheme
	a.decoder = serializer.NewCodecFactory(a.scheme).UniversalDecoder()
	return nil
}

func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	a.kubernetes = clientset
	serverVersion, err := a.kubernetes.Discovery().ServerVersion()
	if err != nil {
		return err
	}
	a.serverVersion = serverVersion

	return nil
}

func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

func (a *actuator) Exists(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure) (bool, error) {
	return infrastructure.Status.LastOperation != nil, nil
}

func (a *actuator) Create(ctx context.Context, config *extensionsv1alpha1.Infrastructure) error {
	return a.reconcile(ctx, config)
}

func (a *actuator) Update(ctx context.Context, config *extensionsv1alpha1.Infrastructure) error {
	return a.reconcile(ctx, config)
}

func (a *actuator) Delete(ctx context.Context, config *extensionsv1alpha1.Infrastructure) error {
	return a.delete(ctx, config)
}

func (a *actuator) reconcile(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure) error {
	infrastructureConfig := &awsapi.InfrastructureConfig{}
	_, _, err := a.decoder.Decode(infrastructure.Spec.ProviderConfig.Raw, nil, infrastructureConfig)
	if err != nil {
		a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "could not decode providerConfig", 10, err)
		return err
	}

	var (
		namespace = infrastructure.Spec.SecretRef.Namespace
		name      = infrastructure.Spec.SecretRef.Name
	)

	providerSecret := &corev1.Secret{}
	if err := a.client.Get(ctx, kutil.Key(namespace, name), providerSecret); err != nil {
		return err
	}

	chartRenderer, err := chartrenderer.New(a.kubernetes)
	if err != nil {
		a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "could not create chart renderer", 30, err)
		return err
	}

	chartValues, err := generateTerraformInfraConfigValues(ctx, infrastructure, infrastructureConfig, providerSecret)
	if err != nil {
		a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "failed to generate terraform config values", 40, err)
		return err
	}

	release, err := chartRenderer.Render(filepath.Join(aws.InternalChartsPath, "aws-infra"), "aws-infra", namespace, chartValues)
	if err != nil {
		a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "could not render chart", 50, err)
		return err
	}

	tf, err := a.getTerraformer(aws.TerrformerPurposeInfra, infrastructure.Namespace, infrastructure.Name)
	if err != nil {
		a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "could not create terraformer object", 70, err)
		return err
	}

	a.recorder.Event(infrastructure, corev1.EventTypeNormal, awsapi.EventReasonCreation, "Applying Terraform config")

	if err := tf.SetVariablesEnvironment(terraformer.GenerateVariablesEnvironment(providerSecret, accessKeysMap)).
		InitializeWith(terraformer.DefaultInitializer(a.client,
			release.FileContent("main.tf"),
			release.FileContent("variables.tf"),
			[]byte(release.FileContent("terraform.tfvars")))).
		Apply(); err != nil {
		a.recorder.Event(infrastructure, corev1.EventTypeWarning, awsapi.EventReasonCreation, fmt.Sprintf("Applying Terraform config %s", err))
		a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "could not apply terraform config", 80, err)
		return err
	}

	if err := a.injectProviderStateIntoStatus(ctx, tf, infrastructure, infrastructureConfig); err != nil {
		a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "failed to update status with the provider state", 90, err)
		return err
	}

	a.recorder.Event(infrastructure, corev1.EventTypeNormal, awsapi.EventReasonCreation, "Infrastructure was provisioned successfully")
	return a.updateStatusSuccess(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "successfully reconciled infrastructure")
}

func (a *actuator) delete(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure) error {
	a.logger.Info("Destroying Infrastructure")
	tf, err := a.getTerraformer(aws.TerrformerPurposeInfra, infrastructure.Namespace, infrastructure.Name)
	if err != nil {
		a.recorder.Event(infrastructure, corev1.EventTypeWarning, awsapi.EventReasonDestruction, fmt.Sprintf("could not create terraformer object: %s", err))
		a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeDelete, "could not create terraformer object", 30, err)
		return err
	}

	configExists, err := tf.ConfigExists()
	if err != nil {
		a.recorder.Event(infrastructure, corev1.EventTypeWarning, awsapi.EventReasonDestruction, fmt.Sprintf("Terraform config does not exist %s", err))
		a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeDelete, "terraform configuration was not found", 60, err)
		return err
	}

	var (
		namespace = infrastructure.Spec.SecretRef.Namespace
		name      = infrastructure.Spec.SecretRef.Name
		region    = infrastructure.Spec.Region
		vpcIDKey  = "vpc_id"
	)

	providerSecret := &corev1.Secret{}
	if err := a.client.Get(ctx, kutil.Key(namespace, name), providerSecret); err != nil {
		return err
	}

	awsClient, err := awsclient.NewClient(string(providerSecret.Data[aws.AccessKeyID]), string(providerSecret.Data[aws.SecretAccessKey]), region)
	if err != nil {
		return err
	}

	var (
		g                                               = flow.NewGraph("AWS infrastructure destruction")
		destroyKubernetesLoadBalancersAndSecurityGroups = g.Add(flow.Task{
			Name: "Destroying Kubernetes load balancers and security groups",
			Fn: flow.SimpleTaskFn(func() error {
				err := a.destroyKubernetesLoadBalancersAndSecurityGroups(ctx, vpcIDKey, namespace, tf, awsClient)
				if err != nil {
					a.recorder.Event(infrastructure, corev1.EventTypeWarning, awsapi.EventReasonDestruction, fmt.Sprintf("failed to destroy kubernetes load balancers and security groups %s", err))
					a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeDelete, "failed to destroy kubernetes load balancers and security groups", 80, err)
					return err
				}
				return nil
			}).RetryUntilTimeout(10*time.Second, 5*time.Minute).DoIf(configExists),
		})

		_ = g.Add(flow.Task{
			Name:         "Destroying Shoot infrastructure",
			Fn:           flow.SimpleTaskFn(tf.SetVariablesEnvironment(terraformer.GenerateVariablesEnvironment(providerSecret, accessKeysMap)).Destroy),
			Dependencies: flow.NewTaskIDs(destroyKubernetesLoadBalancersAndSecurityGroups),
		})

		f = g.Compile()
	)

	a.recorder.Event(infrastructure, corev1.EventTypeNormal, awsapi.EventReasonDestruction, "Destroying Infrastructure")
	err = f.Run(flow.Opts{Logger: gardenerLogger.NewFieldLogger(gardenerLogger.NewLogger("info"), loggerName, "destroyInfrastructure")})
	if err != nil {
		a.recorder.Event(infrastructure, corev1.EventTypeWarning, awsapi.EventReasonDestruction, fmt.Sprintf("failed to run the infrastructure destruction task %s", err))
		a.updateStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeDelete, "failed to run the infrastructure destruction tasks", 70, err)
	}

	a.recorder.Event(infrastructure, corev1.EventTypeNormal, awsapi.EventReasonDestruction, "successfully deleted infrastructure")
	return a.updateStatusSuccess(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeDelete, "successfully deleted infrastructure")
}
