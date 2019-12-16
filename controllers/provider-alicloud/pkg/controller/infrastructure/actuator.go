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
	apisalicloud "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud"
	apisalicloudhelper "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/helper"
	alicloudv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/config"
	confighelper "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/config/helper"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/controller/common"
	extensioncontroller "github.com/gardener/gardener-extensions/pkg/controller"
	controllererrors "github.com/gardener/gardener-extensions/pkg/controller/error"
	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"
	extensionschartrenderer "github.com/gardener/gardener-extensions/pkg/gardener/chartrenderer"
	"github.com/gardener/gardener-extensions/pkg/terraformer"
	"github.com/gardener/gardener-extensions/pkg/util"
	chartutil "github.com/gardener/gardener-extensions/pkg/util/chart"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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
func NewActuator(machineImageMapping []config.MachineImage, machineImageOwnerSecretRef *corev1.SecretReference) infrastructure.Actuator {
	return NewActuatorWithDeps(
		log.Log.WithName("infrastructure-actuator"),
		alicloudclient.NewClientFactory(),
		alicloudclient.DefaultFactory(),
		terraformer.DefaultFactory(),
		extensionschartrenderer.DefaultFactory(),
		DefaultTerraformOps(),
		machineImageMapping,
		machineImageOwnerSecretRef,
	)
}

// NewActuatorWithDeps instantiates an actuator with the given dependencies.
func NewActuatorWithDeps(
	logger logr.Logger,
	newClientFactory alicloudclient.ClientFactory,
	alicloudClientFactory alicloudclient.Factory,
	terraformerFactory terraformer.Factory,
	chartRendererFactory extensionschartrenderer.Factory,
	terraformChartOps TerraformChartOps,
	machineImageMapping []config.MachineImage,
	machineImageOwnerSecretRef *corev1.SecretReference,
) infrastructure.Actuator {
	a := &actuator{
		logger: logger,

		newClientFactory:           newClientFactory,
		alicloudClientFactory:      alicloudClientFactory,
		terraformerFactory:         terraformerFactory,
		chartRendererFactory:       chartRendererFactory,
		terraformChartOps:          terraformChartOps,
		machineImageMapping:        machineImageMapping,
		machineImageOwnerSecretRef: machineImageOwnerSecretRef,
	}

	return a
}

type actuator struct {
	decoder runtime.Decoder
	logger  logr.Logger

	alicloudECSClient     alicloudclient.ECS
	newClientFactory      alicloudclient.ClientFactory
	alicloudClientFactory alicloudclient.Factory
	terraformerFactory    terraformer.Factory
	chartRendererFactory  extensionschartrenderer.Factory
	terraformChartOps     TerraformChartOps

	client client.Client
	config *rest.Config

	chartRenderer chartrenderer.Interface

	machineImageMapping        []config.MachineImage
	machineImageOwnerSecretRef *corev1.SecretReference
}

func (a *actuator) InjectScheme(scheme *runtime.Scheme) error {
	a.decoder = serializer.NewCodecFactory(scheme).UniversalDeserializer()
	return nil
}

// InjectAPIReader implements inject.APIReader and instantiates actuator.alicloudECSClient.
func (a *actuator) InjectAPIReader(reader client.Reader) error {
	if a.machineImageOwnerSecretRef != nil {
		machineImageOwnerSecret := &corev1.Secret{}
		err := reader.Get(context.Background(), client.ObjectKey{
			Name:      a.machineImageOwnerSecretRef.Name,
			Namespace: a.machineImageOwnerSecretRef.Namespace,
		}, machineImageOwnerSecret)
		if err != nil {
			return err
		}
		seedCloudProviderCredentials, err := alicloud.ReadSecretCredentials(machineImageOwnerSecret)
		if err != nil {
			return err
		}
		a.alicloudECSClient, err = a.newClientFactory.NewECSClient(context.Background(), "", seedCloudProviderCredentials.AccessKeyID, seedCloudProviderCredentials.AccessKeySecret)
		return err
	}
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

func (a *actuator) fetchEIPInternetChargeType(vpcClient alicloudclient.VPC, tf terraformer.Terraformer) (string, error) {
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
	tf terraformer.Terraformer,
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

func (a *actuator) newInitializer(infra *extensionsv1alpha1.Infrastructure, config *alicloudv1alpha1.InfrastructureConfig, values *InitializerValues) (terraformer.Initializer, error) {
	chartValues := a.terraformChartOps.ComputeChartValues(infra, config, values)
	release, err := a.chartRenderer.Render(alicloud.InfraChartPath, alicloud.InfraRelease, infra.Namespace, chartValues)
	if err != nil {
		return nil, err
	}

	files, err := chartutil.ExtractTerraformFiles(release)
	if err != nil {
		return nil, err
	}

	terraformState, err := terraformer.UnmarshalRawState(infra.Status.State)
	if err != nil {
		return nil, err
	}

	return a.terraformerFactory.DefaultInitializer(a.client, files.Main, files.Variables, files.TFVars, terraformState.Data), nil
}

func (a *actuator) newTerraformer(infra *extensionsv1alpha1.Infrastructure, credentials *alicloud.Credentials) (terraformer.Terraformer, error) {
	return common.NewTerraformer(a.terraformerFactory, a.config, credentials, TerraformerPurpose, infra.Namespace, infra.Name)
}

func (a *actuator) extractStatus(tf terraformer.Terraformer, infraConfig *alicloudv1alpha1.InfrastructureConfig, machineImages []alicloudv1alpha1.MachineImage) (*alicloudv1alpha1.InfrastructureStatus, error) {
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
		KeyPairName:   vars[TerraformerOutputKeyKeyPairName],
		MachineImages: machineImages,
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

// findMachineImage takes a list of machine images and tries to find the first entry
// whose name and version matches with the given name and version. If no such entry is
// found then an error will be returned.
func findMachineImage(machineImages []alicloudv1alpha1.MachineImage, name, version string) (*alicloudv1alpha1.MachineImage, error) {
	for _, machineImage := range machineImages {
		if machineImage.Name == name && machineImage.Version == version {
			return &machineImage, nil
		}
	}
	return nil, fmt.Errorf("no machine image name %q in version %q found", name, version)
}

func appendMachineImage(machineImages []alicloudv1alpha1.MachineImage, machineImage alicloudv1alpha1.MachineImage) []alicloudv1alpha1.MachineImage {
	if _, err := findMachineImage(machineImages, machineImage.Name, machineImage.Version); err != nil {
		return append(machineImages, machineImage)
	}
	return machineImages
}

// shareCustomizedImages checks whether Shoot's Alicloud account has permissions to use the customized images. If it can't
// access them, these images will be shared with it from Seed's Alicloud account. The list of images that worker use will be
// returned.
func (a *actuator) shareCustomizedImages(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensioncontroller.Cluster) ([]alicloudv1alpha1.MachineImage, error) {
	var (
		machineImages []alicloudv1alpha1.MachineImage
	)

	_, shootCloudProviderCredentials, err := a.getConfigAndCredentialsForInfra(ctx, infra)
	if err != nil {
		return nil, err
	}
	a.logger.Info("Creating Alicloud ECS client for Shoot", "infrastructure", infra.Name)
	shootAlicloudECSClient, err := a.newClientFactory.NewECSClient(ctx, infra.Spec.Region, shootCloudProviderCredentials.AccessKeyID, shootCloudProviderCredentials.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	a.logger.Info("Creating Alicloud STS client for Shoot", "infrastructure", infra.Name)
	shootAlicloudSTSClient, err := a.newClientFactory.NewSTSClient(ctx, infra.Spec.Region, shootCloudProviderCredentials.AccessKeyID, shootCloudProviderCredentials.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	shootCloudProviderAccountID, err := shootAlicloudSTSClient.GetAccountIDFromCallerIdentity(ctx)
	if err != nil {
		return nil, err
	}

	a.logger.Info("Sharing customized image with Shoot's Alicloud account from Seed", "infrastructure", infra.Name)
	for _, worker := range cluster.Shoot.Spec.Provider.Workers {
		imageID, err := confighelper.FindImageForRegion(a.machineImageMapping, worker.Machine.Image.Name, worker.Machine.Image.Version, infra.Spec.Region)
		if err != nil {
			if providerStatus := infra.Status.ProviderStatus; providerStatus != nil {
				infrastructureStatus := &apisalicloud.InfrastructureStatus{}
				if _, _, err := a.decoder.Decode(providerStatus.Raw, nil, infrastructureStatus); err != nil {
					return nil, errors.Wrapf(err, "could not decode infrastructure status of infrastructure '%s'", util.ObjectName(infra))
				}

				machineImage, err := apisalicloudhelper.FindMachineImage(infrastructureStatus.MachineImages, worker.Machine.Image.Name, worker.Machine.Image.Version)
				if err != nil {
					return nil, err
				}
				imageID = machineImage.ID
			} else {
				return nil, err
			}
		}
		machineImages = appendMachineImage(machineImages, alicloudv1alpha1.MachineImage{
			Name:    worker.Machine.Image.Name,
			Version: worker.Machine.Image.Version,
			ID:      imageID,
		})

		exists, err := shootAlicloudECSClient.CheckIfImageExists(ctx, imageID)
		if err != nil {
			return nil, err
		}
		if exists {
			continue
		}
		if a.alicloudECSClient == nil {
			return nil, fmt.Errorf("image sharing is not enabled or configured correctly and Alicloud ECS client is not instantiated in Seed. Please contact Gardener administrator")
		}
		if err := a.alicloudECSClient.ShareImageToAccount(ctx, infra.Spec.Region, imageID, shootCloudProviderAccountID); err != nil {
			return nil, err
		}
	}

	return machineImages, nil
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

	machineImages, err := a.shareCustomizedImages(ctx, infra, cluster)
	if err != nil {
		return errors.Wrapf(err, "failed to share the machine images")
	}

	status, err := a.extractStatus(tf, config, machineImages)
	if err != nil {
		return err
	}

	state, err := tf.GetRawState(ctx)
	if err != nil {
		return err
	}
	stateByte, err := state.Marshal()
	if err != nil {
		return err
	}

	return extensioncontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.client, infra, func() error {
		infra.Status.ProviderStatus = &runtime.RawExtension{Object: status}
		infra.Status.State = &runtime.RawExtension{Raw: stateByte}
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
