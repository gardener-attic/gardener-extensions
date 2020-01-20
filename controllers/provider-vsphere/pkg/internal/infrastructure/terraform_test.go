/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package infrastructure

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-extensions/controllers/provider-vsphere/pkg/apis/vsphere"
	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Terraform", func() {
	var (
		infra              *extensionsv1alpha1.Infrastructure
		cloudProfileConfig *vsphere.CloudProfileConfig
		config             *vsphere.InfrastructureConfig
		networking         corev1beta1.Networking

		dnsServers = []string{"a", "b"}
	)

	BeforeEach(func() {
		config = &vsphere.InfrastructureConfig{}

		infra = &extensionsv1alpha1.Infrastructure{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo",
				Name:      "bar",
			},

			Spec: extensionsv1alpha1.InfrastructureSpec{
				Region: "testregion",
				SecretRef: corev1.SecretReference{
					Namespace: "foo",
					Name:      "vsphere-credentials",
				},
				ProviderConfig: &runtime.RawExtension{
					Object: config,
				},
			},
		}

		cidr := "10.1.0.0/16"
		networking = corev1beta1.Networking{
			Nodes: &cidr,
		}

		dc := "scc01-DC"
		ds := "A800_VMwareB"
		cc := "scc01w01-DEV"
		cloudProfileConfig = &vsphere.CloudProfileConfig{
			NamePrefix: "nameprefix",
			DNSServers: dnsServers,
			Regions: []vsphere.RegionSpec{
				{
					Name:               "testregion",
					VsphereHost:        "vsphere.host.internal",
					VsphereInsecureSSL: true,
					NSXTHost:           "nsxt.host.internal",
					NSXTInsecureSSL:    true,
					TransportZone:      "tz",
					LogicalTier0Router: "lt0router",
					EdgeCluster:        "edgecluster",
					SNATIPPool:         "snatIpPool",
					Datacenter:         &dc,
					Datastore:          &ds,
					Zones: []vsphere.ZoneSpec{
						{
							Name:           "testzone",
							ComputeCluster: &cc,
						},
					},
				},
			},
			Constraints: vsphere.Constraints{
				LoadBalancerConfig: vsphere.LoadBalancerConfig{
					Size: "SMALL",
					Classes: []vsphere.LoadBalancerClass{
						{
							Name:       "default",
							IPPoolName: "lbpool",
						},
					},
				},
			},
			MachineImages: []vsphere.MachineImages{
				{Name: "coreos",
					Versions: []vsphere.MachineImageVersion{
						{
							Version: "2191.5.0",
							Path:    "gardener/templates/coreos-2191.5.0",
							GuestID: "coreos64Guest",
						},
					},
				},
			},
		}
	})

	Describe("#ComputeTerraformerChartValues", func() {
		It("should correctly compute the terraformer chart values", func() {
			values, err := ComputeTerraformerChartValues(infra, config, cloudProfileConfig, networking)
			Expect(err).To(BeNil())

			Expect(values).To(Equal(map[string]interface{}{
				"nsxt": map[string]interface{}{
					"host":               "nsxt.host.internal",
					"insecure":           true,
					"transportZone":      "tz",
					"logicalTier0Router": "lt0router",
					"edgeCluster":        "edgecluster",
					"snatIpPool":         "snatIpPool",
					"namePrefix":         "nameprefix",
					"dnsServers":         dnsServers,
				},
				"sshPublicKey": string(infra.Spec.SSHPublicKey),
				"clusterName":  infra.Namespace,
				"networks": map[string]interface{}{
					"worker": *networking.Nodes,
				},
			}))
		})
	})
})
