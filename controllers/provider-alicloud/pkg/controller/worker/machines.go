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

package worker

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud"
	alicloudapi "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud"
	apisalicloud "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud"
	alicloudapihelper "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/helper"
	"github.com/gardener/gardener-extensions/pkg/controller/worker"
	"github.com/gardener/gardener-extensions/pkg/util"

	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// MachineClassKind yields the name of the Alicloud machine class.
func (w *workerDelegate) MachineClassKind() string {
	return "AlicloudMachineClass"
}

// MachineClassList yields a newly initialized AlicloudMachineClassList object.
func (w *workerDelegate) MachineClassList() runtime.Object {
	return &machinev1alpha1.AlicloudMachineClassList{}
}

// DeployMachineClasses generates and creates the Alicloud specific machine classes.
func (w *workerDelegate) DeployMachineClasses(ctx context.Context) error {
	if w.machineClasses == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return err
		}
	}

	return w.seedChartApplier.ApplyChart(ctx, filepath.Join(alicloud.InternalChartsPath, "machineclass"), w.worker.Namespace, "machineclass", map[string]interface{}{"machineClasses": w.machineClasses}, nil)
}

// GenerateMachineDeployments generates the configuration for the desired machine deployments.
func (w *workerDelegate) GenerateMachineDeployments(ctx context.Context) (worker.MachineDeployments, error) {
	if w.machineDeployments == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return nil, err
		}
	}
	return w.machineDeployments, nil
}

func (w *workerDelegate) generateMachineClassSecretData(ctx context.Context) (map[string][]byte, error) {
	credentials, err := alicloud.ReadCredentialsFromSecretRef(ctx, w.client, &w.worker.Spec.SecretRef)
	if err != nil {
		return nil, err
	}

	return map[string][]byte{
		machinev1alpha1.AlicloudAccessKeyID:     []byte(credentials.AccessKeyID),
		machinev1alpha1.AlicloudAccessKeySecret: []byte(credentials.AccessKeySecret),
	}, nil
}

func (w *workerDelegate) generateMachineConfig(ctx context.Context) error {
	var (
		machineDeployments = worker.MachineDeployments{}
		machineClasses     []map[string]interface{}
		machineImages      []apisalicloud.MachineImage
	)

	machineClassSecretData, err := w.generateMachineClassSecretData(ctx)
	if err != nil {
		return err
	}

	shootVersionMajorMinor, err := util.VersionMajorMinor(w.cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return err
	}

	infrastructureStatus := &alicloudapi.InfrastructureStatus{}
	if _, _, err := w.decoder.Decode(w.worker.Spec.InfrastructureProviderStatus.Raw, nil, infrastructureStatus); err != nil {
		return err
	}

	nodesSecurityGroup, err := alicloudapihelper.FindSecurityGroupByPurpose(infrastructureStatus.VPC.SecurityGroups, alicloudapi.PurposeNodes)
	if err != nil {
		return err
	}

	for _, pool := range w.worker.Spec.Pools {
		zoneLen := len(pool.Zones)

		machineImageID, err := w.findMachineImageForRegion(pool.MachineImage.Name, pool.MachineImage.Version, w.worker.Spec.Region)
		if err != nil {
			return err
		}
		machineImages = appendMachineImage(machineImages, apisalicloud.MachineImage{
			Name:    pool.MachineImage.Name,
			Version: pool.MachineImage.Version,
			ID:      machineImageID,
		})

		volumeSize, err := worker.DiskSize(pool.Volume.Size)
		if err != nil {
			return err
		}

		for zoneIndex, zone := range pool.Zones {
			nodesVSwitch, err := alicloudapihelper.FindVSwitchForPurposeAndZone(infrastructureStatus.VPC.VSwitches, alicloudapi.PurposeNodes, zone)
			if err != nil {
				return err
			}

			systemDisk := map[string]interface{}{
				"size": volumeSize,
			}
			if pool.Volume.Type != nil {
				systemDisk["category"] = *pool.Volume.Type
			}

			machineClassSpec := map[string]interface{}{
				"imageID":                 machineImageID,
				"instanceType":            pool.MachineType,
				"region":                  w.worker.Spec.Region,
				"zoneID":                  zone,
				"securityGroupID":         nodesSecurityGroup.ID,
				"vSwitchID":               nodesVSwitch.ID,
				"systemDisk":              systemDisk,
				"instanceChargeType":      "PostPaid",
				"internetChargeType":      "PayByTraffic",
				"internetMaxBandwidthIn":  5,
				"internetMaxBandwidthOut": 5,
				"spotStrategy":            "NoSpot",
				"tags": map[string]string{
					fmt.Sprintf("kubernetes.io/cluster/%s", w.worker.Namespace):     "1",
					fmt.Sprintf("kubernetes.io/role/worker/%s", w.worker.Namespace): "1",
				},
				"secret": map[string]interface{}{
					"userData": string(pool.UserData),
				},
				"keyPairName": infrastructureStatus.KeyPairName,
			}

			var (
				machineClassSpecHash = worker.MachineClassHash(machineClassSpec, shootVersionMajorMinor)
				deploymentName       = fmt.Sprintf("%s-%s-%s", w.worker.Namespace, pool.Name, zone)
				className            = fmt.Sprintf("%s-%s", deploymentName, machineClassSpecHash)
			)

			machineDeployments = append(machineDeployments, worker.MachineDeployment{
				Name:           deploymentName,
				ClassName:      className,
				SecretName:     className,
				Minimum:        worker.DistributeOverZones(zoneIndex, pool.Minimum, zoneLen),
				Maximum:        worker.DistributeOverZones(zoneIndex, pool.Maximum, zoneLen),
				MaxSurge:       worker.DistributePositiveIntOrPercent(zoneIndex, pool.MaxSurge, zoneLen, pool.Maximum),
				MaxUnavailable: worker.DistributePositiveIntOrPercent(zoneIndex, pool.MaxUnavailable, zoneLen, pool.Minimum),
				Labels:         pool.Labels,
				Annotations:    pool.Annotations,
				Taints:         pool.Taints,
			})

			machineClassSpec["name"] = className
			machineClassSpec["secret"].(map[string]interface{})[alicloud.AccessKeyID] = string(machineClassSecretData[machinev1alpha1.AlicloudAccessKeyID])
			machineClassSpec["secret"].(map[string]interface{})[alicloud.AccessKeySecret] = string(machineClassSecretData[machinev1alpha1.AlicloudAccessKeySecret])

			machineClasses = append(machineClasses, machineClassSpec)
		}
	}

	w.machineDeployments = machineDeployments
	w.machineClasses = machineClasses
	w.machineImages = machineImages

	return nil
}
