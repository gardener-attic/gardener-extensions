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
	"strconv"
	"strings"
	"time"

	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud"
	alicloudclient "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud/client"
	alicloudv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/controller/common"
	extensioncontroller "github.com/gardener/gardener-extensions/pkg/controller"
	controllererrors "github.com/gardener/gardener-extensions/pkg/controller/error"
	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"
	extensionschartrenderer "github.com/gardener/gardener-extensions/pkg/gardener/chartrenderer"
	extensionsterraformer "github.com/gardener/gardener-extensions/pkg/gardener/terraformer"
	"github.com/gardener/gardener-extensions/pkg/terraformer"
	chartutil "github.com/gardener/gardener-extensions/pkg/util/chart"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// StatusTypeMeta is the TypeMeta of InfrastructureStatus.
var StatusTypeMeta = func() metav1.TypeMeta {
	apiVersion, kind := alicloudv1alpha1.SchemeGroupVersion.WithKind(extensioncontroller.UnsafeGuessKind(&alicloudv1alpha1.InfrastructureStatus{})).ToAPIVersionAndKind()
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

func (a *actuator) getConfigAndCredentialsForInfra(ctx context.Context, infra *extensionsv1alpha1.Infrastructure) (*alicloudv1alpha1.InfrastructureConfig, *alicloud.Credentials, error) {
	config := &alicloudv1alpha1.InfrastructureConfig{}
	if _, _, err := a.decoder.Decode(infra.Spec.ProviderConfig.Raw, nil, config); err != nil {
		return nil, nil, err
	}

	credentials, err := alicloud.ReadCredentialsFromSecretRef(ctx, a.client, &infra.Spec.SecretRef)
	if err != nil {
		return nil, nil, err
	}

	return config, credentials, nil
}

func (a *actuator) fetchEIPInternetChargeType(vpcClient alicloudclient.VPC, tf extensionsterraformer.Interface) (string, error) {
	stateVariables, err := tf.GetStateOutputVariables(TerraformerOutputKeyVPCID)
	if err != nil {
		if apierrors.IsNotFound(err) || terraformer.IsVariablesNotFoundError(err) {
			return alicloudclient.DefaultInternetChargeType, nil
		}
		return "", err
	}

	return FetchEIPInternetChargeType(vpcClient, stateVariables[TerraformerOutputKeyVPCID])
}

func (a *actuator) getInitializerValues(
	tf extensionsterraformer.Interface,
	infra *extensionsv1alpha1.Infrastructure,
	config *alicloudv1alpha1.InfrastructureConfig,
	credentials *alicloud.Credentials,
) (*InitializerValues, error) {
	vpcClient, err := a.alicloudClientFactory.NewVPC(infra.Spec.Region, credentials.AccessKeyID, credentials.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	if config.Networks.VPC.ID == nil {
		internetChargeType, err := a.fetchEIPInternetChargeType(vpcClient, tf)
		if err != nil {
			return nil, err
		}

		return a.terraformChartOps.ComputeCreateVPCInitializerValues(config, internetChargeType), nil
	}

	vpcID := *config.Networks.VPC.ID

	vpcInfo, err := GetVPCInfo(vpcClient, vpcID)
	if err != nil {
		return nil, err
	}

	return a.terraformChartOps.ComputeUseVPCInitializerValues(config, vpcInfo), nil
}

func (a *actuator) newInitializer(infra *extensionsv1alpha1.Infrastructure, config *alicloudv1alpha1.InfrastructureConfig, values *InitializerValues) (extensionsterraformer.Initializer, error) {
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

func (a *actuator) newTerraformer(infra *extensionsv1alpha1.Infrastructure, credentials *alicloud.Credentials) (extensionsterraformer.Interface, error) {
	return common.NewTerraformer(a.terraformerFactory, a.config, credentials, TerraformerPurpose, infra.Namespace, infra.Name)
}

func (a *actuator) extractStatus(tf extensionsterraformer.Interface, infraConfig *alicloudv1alpha1.InfrastructureConfig) (*alicloudv1alpha1.InfrastructureStatus, error) {
	outputVarKeys := []string{
		TerraformerOutputKeyVPCID,
		TerraformerOutputKeyVPCCIDR,
		TerraformerOutputKeySecurityGroupID,
		TerraformerOutputKeyKeyPairName,
	}

	for zoneIndex := range infraConfig.Networks.Zones {
		outputVarKeys = append(outputVarKeys, fmt.Sprintf("%s%d", TerraformerOutputKeyVSwitchNodesPrefix, zoneIndex))
	}

	vars, err := tf.GetStateOutputVariables(outputVarKeys...)
	if err != nil {
		return nil, err
	}

	vswitches, err := computeProviderStatusVSwitches(infraConfig, vars)
	if err != nil {
		return nil, err
	}

	return &alicloudv1alpha1.InfrastructureStatus{
		TypeMeta: StatusTypeMeta,
		VPC: alicloudv1alpha1.VPCStatus{
			ID:        vars[TerraformerOutputKeyVPCID],
			VSwitches: vswitches,
			SecurityGroups: []alicloudv1alpha1.SecurityGroup{
				{
					Purpose: alicloudv1alpha1.PurposeNodes,
					ID:      vars[TerraformerOutputKeySecurityGroupID],
				},
			},
		},
		KeyPairName: vars[TerraformerOutputKeyKeyPairName],
	}, nil
}

func computeProviderStatusVSwitches(infrastructure *alicloudv1alpha1.InfrastructureConfig, values map[string]string) ([]alicloudv1alpha1.VSwitch, error) {
	var vswitchesToReturn []alicloudv1alpha1.VSwitch

	for key, value := range values {
		var (
			prefix  string
			purpose alicloudv1alpha1.Purpose
		)

		if strings.HasPrefix(key, TerraformerOutputKeyVSwitchNodesPrefix) {
			prefix = TerraformerOutputKeyVSwitchNodesPrefix
			purpose = alicloudv1alpha1.PurposeNodes
		}

		if len(prefix) == 0 {
			continue
		}

		zoneID, err := strconv.Atoi(strings.TrimPrefix(key, prefix))
		if err != nil {
			return nil, err
		}
		vswitchesToReturn = append(vswitchesToReturn, alicloudv1alpha1.VSwitch{
			ID:      value,
			Purpose: purpose,
			Zone:    infrastructure.Networks.Zones[zoneID].Name,
		})
	}

	return vswitchesToReturn, nil
}

// Reconcile implements infrastructure.Actuator.
func (a *actuator) Reconcile(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensioncontroller.Cluster) error {
	config, credentials, err := a.getConfigAndCredentialsForInfra(ctx, infra)
	if err != nil {
		return err
	}

	tf, err := a.newTerraformer(infra, credentials)
	if err != nil {
		return err
	}

	initializerValues, err := a.getInitializerValues(tf, infra, config, credentials)
	if err != nil {
		return err
	}

	initializer, err := a.newInitializer(infra, config, initializerValues)
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

	status, err := a.extractStatus(tf, config)
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
