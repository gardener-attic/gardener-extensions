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

	azureapi "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/azure"
	azureapihelper "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/azure/helper"
	confighelper "github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/apis/config/helper"
	"github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/azure"
	"github.com/gardener/gardener-extensions/controllers/provider-azure/pkg/internal"
	"github.com/gardener/gardener-extensions/pkg/controller/worker"
	"github.com/gardener/gardener-extensions/pkg/util"

	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// MachineClassKind yields the name of the AWS machine class.
func (w *workerDelegate) MachineClassKind() string {
	return "AzureMachineClass"
}

// MachineClassList yields a newly initialized AzureMachineClassList object.
func (w *workerDelegate) MachineClassList() runtime.Object {
	return &machinev1alpha1.AzureMachineClassList{}
}

// DeployMachineClasses generates and creates the AWS specific machine classes.
func (w *workerDelegate) DeployMachineClasses(ctx context.Context) error {
	if w.machineClasses == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return err
		}
	}
	return w.seedChartApplier.ApplyChart(ctx, filepath.Join(azure.InternalChartsPath, "machineclass"), w.worker.Namespace, "machineclass", map[string]interface{}{"machineClasses": w.machineClasses}, nil)
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
	credentials, err := internal.GetClientAuthData(ctx, w.client, w.worker.Spec.SecretRef)
	if err != nil {
		return nil, err
	}

	return map[string][]byte{
		machinev1alpha1.AzureClientID:       []byte(credentials.ClientID),
		machinev1alpha1.AzureClientSecret:   []byte(credentials.ClientSecret),
		machinev1alpha1.AzureSubscriptionID: []byte(credentials.SubscriptionID),
		machinev1alpha1.AzureTenantID:       []byte(credentials.TenantID),
	}, nil
}

func (w *workerDelegate) generateMachineConfig(ctx context.Context) error {
	var (
		machineDeployments = worker.MachineDeployments{}
		machineClasses     []map[string]interface{}
	)

	machineClassSecretData, err := w.generateMachineClassSecretData(ctx)
	if err != nil {
		return err
	}

	shootVersionMajorMinor, err := util.VersionMajorMinor(w.cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return err
	}

	infrastructureStatus := &azureapi.InfrastructureStatus{}
	if _, _, err := w.decoder.Decode(w.worker.Spec.InfrastructureProviderStatus.Raw, nil, infrastructureStatus); err != nil {
		return err
	}

	nodesSubnet, err := azureapihelper.FindSubnetByPurpose(infrastructureStatus.Networks.Subnets, azureapi.PurposeNodes)
	if err != nil {
		return err
	}
	nodesAvailabilitySet, err := azureapihelper.FindAvailabilitySetByPurpose(infrastructureStatus.AvailabilitySets, azureapi.PurposeNodes)
	if err != nil {
		return err
	}

	for _, pool := range w.worker.Spec.Pools {
		machineImage, err := confighelper.FindImage(w.machineImages, pool.MachineImage.Name, pool.MachineImage.Version)
		if err != nil {
			return err
		}

		volumeSize, err := worker.DiskSize(pool.Volume.Size)
		if err != nil {
			return err
		}

		machineClassSpec := map[string]interface{}{
			"region":            w.worker.Spec.Region,
			"resourceGroup":     infrastructureStatus.ResourceGroup.Name,
			"vnetName":          infrastructureStatus.Networks.VNet.Name,
			"subnetName":        nodesSubnet.Name,
			"availabilitySetID": nodesAvailabilitySet.ID,
			"tags": map[string]interface{}{
				"Name": w.worker.Namespace,
				fmt.Sprintf("kubernetes.io-cluster-%s", w.worker.Namespace): "1",
				"kubernetes.io-role-node":                                   "1",
			},
			"secret": map[string]interface{}{
				"cloudConfig": string(pool.UserData),
			},
			"machineType": pool.MachineType,
			"image": map[string]interface{}{
				"publisher": machineImage.Publisher,
				"offer":     machineImage.Offer,
				"sku":       machineImage.SKU,
				"version":   machineImage.Version,
			},
			"volumeSize":   volumeSize,
			"sshPublicKey": string(w.worker.Spec.SSHPublicKey),
		}

		var (
			machineClassSpecHash = worker.MachineClassHash(machineClassSpec, shootVersionMajorMinor)
			deploymentName       = fmt.Sprintf("%s-%s", w.worker.Namespace, pool.Name)
			className            = fmt.Sprintf("%s-%s", deploymentName, machineClassSpecHash)
		)

		machineDeployments = append(machineDeployments, worker.MachineDeployment{
			Name:           deploymentName,
			ClassName:      className,
			SecretName:     className,
			Minimum:        pool.Minimum,
			Maximum:        pool.Maximum,
			MaxSurge:       pool.MaxSurge,
			MaxUnavailable: pool.MaxUnavailable,
			Labels:         pool.Labels,
			Annotations:    pool.Annotations,
			Taints:         pool.Taints,
		})

		machineClassSpec["name"] = className
		machineClassSpec["secret"].(map[string]interface{})[internal.ClientIDKey] = string(machineClassSecretData[machinev1alpha1.AzureClientID])
		machineClassSpec["secret"].(map[string]interface{})[internal.ClientSecretKey] = string(machineClassSecretData[machinev1alpha1.AzureClientSecret])
		machineClassSpec["secret"].(map[string]interface{})[internal.SubscriptionIDKey] = string(machineClassSecretData[machinev1alpha1.AzureSubscriptionID])
		machineClassSpec["secret"].(map[string]interface{})[internal.TenantIDKey] = string(machineClassSecretData[machinev1alpha1.AzureTenantID])

		machineClasses = append(machineClasses, machineClassSpec)
	}

	w.machineDeployments = machineDeployments
	w.machineClasses = machineClasses

	return nil
}
