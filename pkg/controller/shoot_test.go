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
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Shoot", func() {
	trueVar := true
	falseVar := false

	cidr := "10.250.0.0/19"

	DescribeTable("#GetPodNetwork",
		func(cluster *Cluster, cidr string) {
			Expect(GetPodNetwork(cluster)).To(Equal(cidr))
		},

		Entry("pod cidr is given", &Cluster{
			CoreShoot: &gardencorev1alpha1.Shoot{
				Spec: gardencorev1alpha1.ShootSpec{
					Networking: gardencorev1alpha1.Networking{
						Pods: &cidr,
					},
				},
			},
		}, cidr),

		Entry("cloud is AWS", &Cluster{
			Shoot: &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						AWS: &gardenv1beta1.AWSCloud{
							Networks: gardenv1beta1.AWSNetworks{
								K8SNetworks: gardenv1beta1.K8SNetworks{Pods: &cidr},
							},
						},
					},
				},
			},
		}, cidr),

		Entry("cloud is Azure", &Cluster{
			Shoot: &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						Azure: &gardenv1beta1.AzureCloud{
							Networks: gardenv1beta1.AzureNetworks{
								K8SNetworks: gardenv1beta1.K8SNetworks{Pods: &cidr},
							},
						},
					},
				},
			},
		}, cidr),

		Entry("cloud is GCP", &Cluster{
			Shoot: &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						GCP: &gardenv1beta1.GCPCloud{
							Networks: gardenv1beta1.GCPNetworks{
								K8SNetworks: gardenv1beta1.K8SNetworks{Pods: &cidr},
							},
						},
					},
				},
			},
		}, cidr),

		Entry("cloud is OpenStack", &Cluster{
			Shoot: &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						OpenStack: &gardenv1beta1.OpenStackCloud{
							Networks: gardenv1beta1.OpenStackNetworks{
								K8SNetworks: gardenv1beta1.K8SNetworks{Pods: &cidr},
							},
						},
					},
				},
			},
		}, cidr),

		Entry("cloud is Alicloud", &Cluster{
			Shoot: &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						Alicloud: &gardenv1beta1.Alicloud{
							Networks: gardenv1beta1.AlicloudNetworks{
								K8SNetworks: gardenv1beta1.K8SNetworks{Pods: &cidr},
							},
						},
					},
				},
			},
		}, cidr),

		Entry("cloud is Packet", &Cluster{
			Shoot: &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						Packet: &gardenv1beta1.PacketCloud{
							Networks: gardenv1beta1.PacketNetworks{
								K8SNetworks: gardenv1beta1.K8SNetworks{Pods: &cidr},
							},
						},
					},
				},
			},
		}, cidr),
	)

	DescribeTable("#IsHibernated (gardenv1beta1.Shoot)",
		func(hibernation *gardenv1beta1.Hibernation, expectation bool) {
			cluster := &Cluster{
				Shoot: &gardenv1beta1.Shoot{
					Spec: gardenv1beta1.ShootSpec{
						Hibernation: hibernation,
					},
				},
			}

			Expect(IsHibernated(cluster)).To(Equal(expectation))
		},

		Entry("hibernation is nil", nil, false),
		Entry("hibernation is not enabled", &gardenv1beta1.Hibernation{Enabled: &falseVar}, false),
		Entry("hibernation is enabled", &gardenv1beta1.Hibernation{Enabled: &trueVar}, true),
	)

	DescribeTable("#IsHibernated (gardencorev1alpha1.Shoot)",
		func(hibernation *gardencorev1alpha1.Hibernation, expectation bool) {
			cluster := &Cluster{
				CoreShoot: &gardencorev1alpha1.Shoot{
					Spec: gardencorev1alpha1.ShootSpec{
						Hibernation: hibernation,
					},
				},
			}

			Expect(IsHibernated(cluster)).To(Equal(expectation))
		},

		Entry("hibernation is nil", nil, false),
		Entry("hibernation is not enabled", &gardencorev1alpha1.Hibernation{Enabled: &falseVar}, false),
		Entry("hibernation is enabled", &gardencorev1alpha1.Hibernation{Enabled: &trueVar}, true),
	)

	DescribeTable("#GetReplicas (gardenv1beta1.Shoot)",
		func(hibernation *gardenv1beta1.Hibernation, wokenUp, expectation int) {
			cluster := &Cluster{
				Shoot: &gardenv1beta1.Shoot{
					Spec: gardenv1beta1.ShootSpec{
						Hibernation: hibernation,
					},
				},
			}

			Expect(GetReplicas(cluster, wokenUp)).To(Equal(expectation))
		},

		Entry("hibernation is not enabled", nil, 3, 3),
		Entry("hibernation is enabled", &gardenv1beta1.Hibernation{Enabled: &trueVar}, 1, 0),
	)

	DescribeTable("#GetReplicas (gardencorev1alpha1.Shoot)",
		func(hibernation *gardencorev1alpha1.Hibernation, wokenUp, expectation int) {
			cluster := &Cluster{
				CoreShoot: &gardencorev1alpha1.Shoot{
					Spec: gardencorev1alpha1.ShootSpec{
						Hibernation: hibernation,
					},
				},
			}

			Expect(GetReplicas(cluster, wokenUp)).To(Equal(expectation))
		},

		Entry("hibernation is not enabled", nil, 3, 3),
		Entry("hibernation is enabled", &gardencorev1alpha1.Hibernation{Enabled: &trueVar}, 1, 0),
	)
})
