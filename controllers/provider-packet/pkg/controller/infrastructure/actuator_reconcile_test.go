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

package infrastructure_test

import (
	. "github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/controller/infrastructure"
	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/packet"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Actuator Reconcile", func() {
	Describe("#GenerateTerraformInfraConfig", func() {
		It("should compute the correct Terraform config", func() {
			var (
				projectID   = "project-1234"
				sshKey      = "foo-bar"
				clusterName = "shoot--foo-bar"

				infrastructure = &extensionsv1alpha1.Infrastructure{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infra",
						Namespace: clusterName,
					},
					Spec: extensionsv1alpha1.InfrastructureSpec{
						SSHPublicKey: []byte(sshKey),
					},
				}
			)

			Expect(GenerateTerraformInfraConfig(infrastructure, projectID)).To(Equal(map[string]interface{}{
				"packet": map[string]interface{}{
					"projectID": projectID,
				},
				"sshPublicKey": sshKey,
				"clusterName":  clusterName,
				"outputKeys": map[string]interface{}{
					"sshKeyID": packet.SSHKeyID,
				},
			}))
		})
	})
})
