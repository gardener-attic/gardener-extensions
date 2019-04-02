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
	"testing"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = Describe("Shoot", func() {

	Describe("#GetPodNetwork", func() {
		cidr := gardencorev1alpha1.CIDR("10.250.0.0/19")
		It("should return the AWS pod network for a AWS Shoot", func() {
			shoot := &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						AWS: &gardenv1beta1.AWSCloud{
							Networks: gardenv1beta1.AWSNetworks{
								K8SNetworks: gardencorev1alpha1.K8SNetworks{
									Pods: &cidr,
								},
							},
						},
					},
				},
			}
			Expect(GetPodNetwork(shoot)).To(Equal(cidr))
		})
		It("should return the Azure pod network for a Azure Shoot", func() {
			shoot := &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						Azure: &gardenv1beta1.AzureCloud{
							Networks: gardenv1beta1.AzureNetworks{
								K8SNetworks: gardencorev1alpha1.K8SNetworks{
									Pods: &cidr,
								},
							},
						},
					},
				},
			}
			Expect(GetPodNetwork(shoot)).To(Equal(cidr))
		})
	})

	Describe("#GetReplicas", func() {
		It("should return the given number of replicas for a non-hibernated Shoot", func() {
			Expect(GetReplicas(&gardenv1beta1.Shoot{}, 3)).To(Equal(3))
		})
		It("should return 0 for a hibernated Shoot", func() {
			shoot := &gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Hibernation: &gardenv1beta1.Hibernation{
						Enabled: true,
					},
				},
			}
			Expect(GetReplicas(shoot, 1)).To(Equal(0))
		})
	})
})
