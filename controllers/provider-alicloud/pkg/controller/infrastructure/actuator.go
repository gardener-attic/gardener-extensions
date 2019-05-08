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
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud"
	alicloudclient "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud/client"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/controller/common"
	extensioncontroller "github.com/gardener/gardener-extensions/pkg/controller"
	controllererrors "github.com/gardener/gardener-extensions/pkg/controller/error"
	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"
	extensionschartrenderer "github.com/gardener/gardener-extensions/pkg/gardener/chartrenderer"
	extensionsterraformer "github.com/gardener/gardener-extensions/pkg/gardener/terraformer"
	chartutil "github.com/gardener/gardener-extensions/pkg/util/chart"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"time"
)

// StatusTypeMeta is the TypeMeta of InfrastructureStatus.
var StatusTypeMeta = func() metav1.TypeMeta {
	apiVersion, kind := v1alpha1.SchemeGroupVersion.WithKind(extensioncontroller.UnsafeGuessKind(&v1alpha1.InfrastructureStatus{})).ToAPIVersionAndKind()
	return metav1.TypeMeta{
		APIVersion: apiVersion,
		Kind:       kind,
	}
}()

// NewActuator instantiates an actuator with the default dependencies.
func NewActuator() infrastructure.Actuator {
	return NewActuatorWithDeps(
		log.Log.WithName("infrastructure-actuator"),
		alicloudclient.DefaultFactory(),
		extensionsterraformer.DefaultFactory(),
		extensionschartrenderer.DefaultFactory(),
		DefaultTerraformOps(),
	)
}

// NewActuatorWithDeps instantiates an actuator with the given dependencies.
func NewActuatorWithDeps(
	logger logr.Logger,
	alicloudClientFactory alicloudclient.Factory,
	terraformerFactory extensionsterraformer.Factory,
	chartRendererFactory extensionschartrenderer.Factory,
	terraformChartOps TerraformChartOps,
) infrastructure.Actuator {
	a := &actuator{
		logger: logger,

		alicloudClientFactory: alicloudClientFactory,
		terraformerFactory:    terraformerFactory,
		chartRendererFactory:  chartRendererFactory,
		terraformChartOps:     terraformChartOps,
	}

	return a
}

type actuator struct {
	decoder runtime.Decoder
	logger  logr.Logger

	alicloudClientFactory alicloudclient.Factory
	terraformerFactory    extensionsterraformer.Factory
	chartRendererFactory  extensionschartrenderer.Factory
	terraformChartOps     TerraformChartOps

	client client.Client
	config *rest.Config

	chartRenderer chartrenderer.Interface
}

func (a *actuator) InjectScheme(scheme *runtime.Scheme) error {
	a.decoder = serializer.NewCodecFactory(scheme).UniversalDeserializer()
	return nil
}

// InjectClient implements inject.Client.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

// InjectConfig implements inject.Config.
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config

	var err error
	a.chartRenderer, err = a.chartRendererFactory.NewForConfig(config)
	return err
}

func (a *actuator) getConfigAndCredentialsForInfra(ctx context.Context, infra *extensionsv1alpha1.Infrastructure) (*v1alpha1.InfrastructureConfig, *alicloudclient.Credentials, error) {
	config := &v1alpha1.InfrastructureConfig{}
	if _, _, err := a.decoder.Decode(infra.Spec.ProviderConfig.Raw, nil, config); err != nil {
		return nil, nil, err
	}

	secret, err := extensioncontroller.GetSecretByReference(ctx, a.client, &infra.Spec.SecretRef)
	if err != nil {
		return nil, nil, err
	}

	credentials, err := alicloudclient.ReadSecretCredentials(secret)
	if err != nil {
		return nil, nil, err
	}

	return config, credentials, nil
}

func (a *actuator) getInitializerValues(
	infra *extensionsv1alpha1.Infrastructure,
	config *v1alpha1.InfrastructureConfig,
	credentials *alicloudclient.Credentials,
) (*InitializerValues, error) {
	if config.Networks.VPC.ID == nil {
		return a.terraformChartOps.ComputeCreateVPCInitializerValues(config), nil
	}

	vpcID := *config.Networks.VPC.ID
	vpcClient, err := a.alicloudClientFactory.NewVPC(infra.Spec.Region, credentials.AccessKeyID, credentials.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	vpcInfo, err := GetVPCInfo(vpcClient, vpcID)
	if err != nil {
		return nil, err
	}

	return a.terraformChartOps.ComputeUseVPCInitializerValues(config, vpcInfo), nil
}

func (a *actuator) newInitializer(infra *extensionsv1alpha1.Infrastructure, config *v1alpha1.InfrastructureConfig, values *InitializerValues) (extensionsterraformer.Initializer, error) {
	chartValues := a.terraformChartOps.ComputeChartValues(infra, config, values)
	release, err := a.chartRenderer.Render(alicloud.InfraChartPath, alicloud.InfraRelease, infra.Namespace, chartValues)
	if err != nil {
		return nil, err
	}

	files, err := chartutil.ExtractTerraformFiles(release)
	if err != nil {
		return nil, err
	}

	return a.terraformerFactory.DefaultInitializer(a.client, files.Main, files.Variables, files.TFVars), nil
}

func (a *actuator) newTerraformer(infra *extensionsv1alpha1.Infrastructure, credentials *alicloudclient.Credentials) (extensionsterraformer.Interface, error) {
	return common.NewTerraformer(a.terraformerFactory, a.config, credentials, TerraformerPurpose, infra.Namespace, infra.Name)
}

func (a *actuator) extractStatus(tf extensionsterraformer.Interface) (*v1alpha1.InfrastructureStatus, error) {
	vars, err := tf.GetStateOutputVariables(TerraformerOutputKeyVPCID, TerraformerOutputKeyVPCCIDR, TerraformerOutputKeySecurityGroupID, TerraformerOutputKeyKeyPairName)
	if err != nil {
		return nil, err
	}

	var (
		vpcID           = vars[TerraformerOutputKeyVPCID]
		vpcCIDR         = gardencorev1alpha1.CIDR(vars[TerraformerOutputKeyVPCCIDR])
		securityGroupID = vars[TerraformerOutputKeySecurityGroupID]
		keyPairName     = vars[TerraformerOutputKeyKeyPairName]
	)

	return &v1alpha1.InfrastructureStatus{
		TypeMeta: StatusTypeMeta,
		VPC: v1alpha1.VPC{
			ID:   &vpcID,
			CIDR: &vpcCIDR,
		},
		KeyPairName:     keyPairName,
		SecurityGroupID: securityGroupID,
	}, nil
}

// Reconcile implements infrastructure.Actuator.
func (a *actuator) Reconcile(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensioncontroller.Cluster) error {
	config, credentials, err := a.getConfigAndCredentialsForInfra(ctx, infra)
	if err != nil {
		return err
	}

	initializerValues, err := a.getInitializerValues(infra, config, credentials)
	if err != nil {
		return err
	}

	initializer, err := a.newInitializer(infra, config, initializerValues)
	if err != nil {
		return err
	}

	tf, err := a.newTerraformer(infra, credentials)
	if err != nil {
		return err
	}

	if err := tf.InitializeWith(initializer).Apply(); err != nil {
		a.logger.Error(err, "failed to apply the terraform config", "infrastructure", infra.Name)
		return &controllererrors.RequeueAfterError{
			Cause:        err,
			RequeueAfter: 30 * time.Second,
		}
	}

	status, err := a.extractStatus(tf)
	if err != nil {
		return err
	}

	return extensioncontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.client, infra, func() error {
		infra.Status.ProviderStatus = &runtime.RawExtension{Object: status}
		return nil
	})
}

// Delete implements infrastructure.Actuator.
func (a *actuator) Delete(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensioncontroller.Cluster) error {
	_, credentials, err := a.getConfigAndCredentialsForInfra(ctx, infra)
	if err != nil {
		return err
	}

	tf, err := a.newTerraformer(infra, credentials)
	if err != nil {
		return err
	}

	configExists, err := tf.ConfigExists()
	if err != nil {
		return err
	}
	if !configExists {
		return nil
	}

	return tf.Destroy()
}
