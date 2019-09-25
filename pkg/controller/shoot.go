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

package controller

import (
	"github.com/gardener/gardener/pkg/chartrenderer"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// ShootNoCleanupLabel is a constant for a label on a resource indicating the the Gardener cleaner should not delete this
	// resource when cleaning a shoot during the deletion flow.
	ShootNoCleanupLabel = "shoot.gardener.cloud/no-cleanup"
)

// ChartRendererFactory creates chartrenderer.Interface to be used by this actuator.
type ChartRendererFactory interface {
	// NewChartRendererForShoot creates a new chartrenderer.Interface for the shoot cluster.
	NewChartRendererForShoot(string) (chartrenderer.Interface, error)
}

// ChartRendererFactoryFunc is a function that satisfies ChartRendererFactory.
type ChartRendererFactoryFunc func(string) (chartrenderer.Interface, error)

// NewChartRendererForShoot creates a new chartrenderer.Interface for the shoot cluster.
func (f ChartRendererFactoryFunc) NewChartRendererForShoot(version string) (chartrenderer.Interface, error) {
	return f(version)
}

// GetPodNetwork returns the pod network CIDR of the given Shoot.
func GetPodNetwork(cluster *Cluster) string {
	if cluster.Shoot != nil {
		cloud := cluster.Shoot.Spec.Cloud
		switch {
		case cloud.AWS != nil:
			return *cloud.AWS.Networks.K8SNetworks.Pods
		case cloud.Azure != nil:
			return *cloud.Azure.Networks.K8SNetworks.Pods
		case cloud.GCP != nil:
			return *cloud.GCP.Networks.K8SNetworks.Pods
		case cloud.OpenStack != nil:
			return *cloud.OpenStack.Networks.K8SNetworks.Pods
		case cloud.Alicloud != nil:
			return *cloud.Alicloud.Networks.K8SNetworks.Pods
		case cloud.Packet != nil:
			return *cloud.Packet.Networks.K8SNetworks.Pods
		default:
			return ""
		}
	} else if cluster.CoreShoot != nil && cluster.CoreShoot.Spec.Networking.Pods != nil {
		return *cluster.CoreShoot.Spec.Networking.Pods
	}
	return ""
}

// GetServiceNetwork returns the service network CIDR of the given Shoot.
func GetServiceNetwork(cluster *Cluster) string {
	if cluster.Shoot != nil {
		cloud := cluster.Shoot.Spec.Cloud
		switch {
		case cloud.AWS != nil:
			return *cloud.AWS.Networks.K8SNetworks.Services
		case cloud.Azure != nil:
			return *cloud.Azure.Networks.K8SNetworks.Services
		case cloud.GCP != nil:
			return *cloud.GCP.Networks.K8SNetworks.Services
		case cloud.OpenStack != nil:
			return *cloud.OpenStack.Networks.K8SNetworks.Services
		case cloud.Alicloud != nil:
			return *cloud.Alicloud.Networks.K8SNetworks.Services
		case cloud.Packet != nil:
			return *cloud.Packet.Networks.K8SNetworks.Services
		default:
			return ""
		}
	} else if cluster.CoreShoot != nil && cluster.CoreShoot.Spec.Networking.Services != nil {
		return *cluster.CoreShoot.Spec.Networking.Services
	}
	return ""
}

// IsHibernated returns true if the shoot is hibernated, or false otherwise.
func IsHibernated(cluster *Cluster) bool {
	if cluster.Shoot != nil {
		return cluster.Shoot.Spec.Hibernation != nil && cluster.Shoot.Spec.Hibernation.Enabled != nil && *cluster.Shoot.Spec.Hibernation.Enabled
	} else if cluster.CoreShoot != nil {
		return cluster.CoreShoot.Spec.Hibernation != nil && cluster.CoreShoot.Spec.Hibernation.Enabled != nil && *cluster.CoreShoot.Spec.Hibernation.Enabled
	}
	return false
}

// GetKubernetesVersion returns the kubernetes version.
func GetKubernetesVersion(cluster *Cluster) string {
	if cluster.Shoot != nil {
		return cluster.Shoot.Spec.Kubernetes.Version
	} else if cluster.CoreShoot != nil {
		return cluster.CoreShoot.Spec.Kubernetes.Version
	}
	return ""
}

// GetTechnicalID returns the shoot technical id.
func GetTechnicalID(cluster *Cluster) string {
	if cluster.Shoot != nil {
		return cluster.Shoot.Status.TechnicalID
	} else if cluster.CoreShoot != nil {
		return cluster.CoreShoot.Status.TechnicalID
	}
	return ""
}

// GetUID returns the shoot uid.
func GetUID(cluster *Cluster) types.UID {
	if cluster.Shoot != nil {
		return cluster.Shoot.Status.UID
	} else if cluster.CoreShoot != nil {
		return cluster.CoreShoot.Status.UID
	}
	return ""
}

// GetReplicas returns the woken up replicas of the given Shoot.
func GetReplicas(cluster *Cluster, wokenUp int) int {
	if IsHibernated(cluster) {
		return 0
	}
	return wokenUp
}

// GetControlPlaneReplicas returns the woken up replicas for controlplane components of the given Shoot
// that should only be scaled down at the end of the flow.
func GetControlPlaneReplicas(cluster *Cluster, scaledDown bool, wokenUp int) int {
	if ((cluster.Shoot != nil && cluster.Shoot.DeletionTimestamp == nil) ||
		(cluster.CoreShoot != nil && cluster.CoreShoot.DeletionTimestamp == nil)) &&
		IsHibernated(cluster) && scaledDown {
		return 0
	}
	return wokenUp
}
