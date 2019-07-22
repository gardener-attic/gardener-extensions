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

package app

import (
	np "github.com/gardener/gardener-extensions/test/e2e/framework/networkpolicies"
	"github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
)

// alicloudNetworkPolicy holds alicloud-specific network policy settings.
type alicloudNetworkPolicy struct {
	np.Agnostic

	// cloudControllerManagerNotSecured points to alicloud-specific cloud-controller-manager.
	// For now it listens only on HTTP for all Shoot versions.
	cloudControllerManagerNotSecured *np.SourcePod

	// kubeControllerManagerSecured points to alicloud-specific kube-controller-manager.
	kubeControllerManagerSecured *np.SourcePod

	// kubeControllerManagerNotSecured points to alicloud-specific kube-controller-manager.
	kubeControllerManagerNotSecured *np.SourcePod

	// csiPlugin points to alicloud-specific CSI Plugin.
	csiPlugin *np.SourcePod

	// metadata points to alicloud-specific Metadata service.
	metadata *np.Host
}

// NewCloudAware returns alicloud Pod.
func NewCloudAware() np.CloudAware {
	return &alicloudNetworkPolicy{

		cloudControllerManagerNotSecured: &np.SourcePod{
			Ports: np.NewSinglePort(10253),
			Pod: np.NewPod("cloud-controller-manager-http", labels.Set{
				"app":                     "kubernetes",
				"garden.sapcloud.io/role": "controlplane",
				"role":                    "cloud-controller-manager",
			}),
			ExpectedPolicies: sets.NewString(
				"allow-from-prometheus",
				"allow-to-dns",
				"allow-to-public-networks",
				"allow-to-shoot-apiserver",
				"deny-all",
			),
		},

		kubeControllerManagerSecured: &np.SourcePod{
			Ports: np.NewSinglePort(10257),
			Pod: np.NewPod("kube-controller-manager-https", labels.Set{
				"app":                     "kubernetes",
				"garden.sapcloud.io/role": "controlplane",
				"role":                    "controller-manager",
			}, ">= 1.13"),
			ExpectedPolicies: sets.NewString(
				"allow-from-prometheus",
				"allow-to-dns",
				"allow-to-shoot-apiserver",
				"deny-all",
			),
		},

		kubeControllerManagerNotSecured: &np.SourcePod{
			Ports: np.NewSinglePort(10252),
			Pod: np.NewPod("kube-controller-manager-http", labels.Set{
				"app":                     "kubernetes",
				"garden.sapcloud.io/role": "controlplane",
				"role":                    "controller-manager",
			}, "< 1.13"),
			ExpectedPolicies: sets.NewString(
				"allow-from-prometheus",
				"allow-to-dns",
				"allow-to-shoot-apiserver",
				"deny-all",
			),
		},

		csiPlugin: &np.SourcePod{
			Ports: np.NewSinglePort(80),
			Pod: np.NewPod("csi-plugin-controller", labels.Set{
				"app":                     "kubernetes",
				"garden.sapcloud.io/role": "controlplane",
				"role":                    "csi-plugin-controller",
			}),
			ExpectedPolicies: sets.NewString(
				"allow-to-public-networks",
				"allow-to-dns",
				"allow-to-shoot-apiserver",
				"deny-all",
			),
		},

		metadata: &np.Host{
			Description: "Metadata service",
			HostName:    "100.100.100.200",
			Port:        80,
		},
	}

}

// Sources returns list of all alicloud-specific sources and targets.
func (a *alicloudNetworkPolicy) Rules() []np.Rule {
	ag := a.Agnostic
	return []np.Rule{
		a.newSource(a.cloudControllerManagerNotSecured).AllowPod(ag.KubeAPIServer()).AllowHost(ag.External()).Build(),
		a.newSource(a.csiPlugin).AllowPod(ag.KubeAPIServer()).AllowHost(ag.External()).Build(),
		a.newSource(a.kubeControllerManagerSecured).AllowPod(ag.KubeAPIServer()).Build(),
		a.newSource(a.kubeControllerManagerNotSecured).AllowPod(ag.KubeAPIServer()).Build(),
		a.newSource(ag.KubeAPIServer()).AllowPod(ag.EtcdMain(), ag.EtcdEvents()).AllowHost(ag.SeedKubeAPIServer(), ag.External()).Build(),
		a.newSource(ag.EtcdMain()).AllowHost(ag.External()).Build(),
		a.newSource(ag.EtcdEvents()).AllowHost(ag.External()).Build(),
		a.newSource(ag.DependencyWatchdog()).AllowHost(ag.SeedKubeAPIServer(), ag.External()).Build(),
		a.newSource(ag.ElasticSearch()).Build(),
		a.newSource(ag.Grafana()).AllowPod(ag.Prometheus()).Build(),
		a.newSource(ag.Kibana()).AllowTargetPod(ag.ElasticSearch().FromPort("http")).Build(),
		a.newSource(ag.AddonManager()).AllowPod(ag.KubeAPIServer()).AllowHost(ag.SeedKubeAPIServer(), ag.External()).Build(),
		a.newSource(ag.KubeSchedulerNotSecured()).AllowPod(ag.KubeAPIServer()).Build(),
		a.newSource(ag.KubeSchedulerSecured()).AllowPod(ag.KubeAPIServer()).Build(),
		a.newSource(ag.KubeStateMetricsShoot()).AllowPod(ag.KubeAPIServer()).Build(),
		a.newSource(ag.KubeStateMetricsSeed()).AllowHost(ag.SeedKubeAPIServer(), ag.External()).Build(),
		a.newSource(ag.MachineControllerManager()).AllowPod(ag.KubeAPIServer()).AllowHost(ag.SeedKubeAPIServer(), ag.External()).Build(),
		a.newSource(ag.Prometheus()).AllowPod(
			a.cloudControllerManagerNotSecured,
			a.kubeControllerManagerNotSecured,
			a.kubeControllerManagerSecured,
			ag.EtcdEvents(),
			ag.EtcdMain(),
			ag.KubeAPIServer(),
			ag.KubeSchedulerNotSecured(),
			ag.KubeSchedulerSecured(),
			ag.KubeStateMetricsSeed(),
			ag.KubeStateMetricsShoot(),
			ag.MachineControllerManager(),
		).AllowTargetPod(ag.ElasticSearch().FromPort("metrics")).AllowHost(ag.SeedKubeAPIServer(), ag.External(), ag.GardenPrometheus()).Build(),
	}
}

// EgressFromOtherNamespaces returns list of all alicloud-specific sources and targets.
func (a *alicloudNetworkPolicy) EgressFromOtherNamespaces(sourcePod *np.SourcePod) np.Rule {
	ag := a.Agnostic
	return np.NewSource(sourcePod).DenyPod(a.Sources()...).AllowPod(ag.KubeAPIServer()).Build()
}

func (a *alicloudNetworkPolicy) newSource(sourcePod *np.SourcePod) *np.RuleBuilder {
	ag := a.Agnostic
	return np.NewSource(sourcePod).DenyPod(a.Sources()...).DenyHost(a.metadata, ag.External(), ag.GardenPrometheus())
}

// Sources returns a list of SourcePods of AliCloud.
func (a *alicloudNetworkPolicy) Sources() []*np.SourcePod {
	ag := a.Agnostic
	return []*np.SourcePod{
		a.cloudControllerManagerNotSecured,
		a.csiPlugin,
		a.kubeControllerManagerNotSecured,
		a.kubeControllerManagerSecured,
		ag.AddonManager(),
		ag.DependencyWatchdog(),
		ag.ElasticSearch(),
		ag.EtcdEvents(),
		ag.EtcdMain(),
		ag.Grafana(),
		ag.Kibana(),
		ag.KubeAPIServer(),
		ag.KubeSchedulerNotSecured(),
		ag.KubeSchedulerSecured(),
		ag.KubeStateMetricsSeed(),
		ag.KubeStateMetricsShoot(),
		ag.MachineControllerManager(),
		ag.Prometheus(),
	}
}

// Provider returns Alicloud cloud provider.
func (a *alicloudNetworkPolicy) Provider() v1beta1.CloudProvider {
	return v1beta1.CloudProviderAlicloud
}
