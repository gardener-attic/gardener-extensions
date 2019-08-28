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

package common

import (
	"time"

	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/imagevector"
	mockterraformer "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/gardener/terraformer"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
)

var _ = Describe("Terraform", func() {
	var (
		ctrl *gomock.Controller
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#NewTerraformer", func() {
		It("should create a new terraformer and initialize it with the credentials", func() {
			var (
				factory         = mockterraformer.NewMockFactory(ctrl)
				tf              = mockterraformer.NewMockInterface(ctrl)
				config          rest.Config
				accessKeyID     = "accessKeyID"
				accessKeySecret = "accessKeySecret"
				credentials     = alicloud.Credentials{
					AccessKeyID:     accessKeyID,
					AccessKeySecret: accessKeySecret,
				}
				purpose   = "purpose"
				namespace = "namespace"
				name      = "name"
			)

			gomock.InOrder(
				factory.EXPECT().
					NewForConfig(gomock.Any(), &config, purpose, namespace, name, imagevector.TerraformerImage()).
					Return(tf, nil),
				tf.EXPECT().SetVariablesEnvironment(map[string]string{
					TerraformVarAccessKeyID:     accessKeyID,
					TerraformVarAccessKeySecret: accessKeySecret,
				}).Return(tf),
				tf.EXPECT().SetJobBackoffLimit(int32(1)).Return(tf),
				tf.EXPECT().SetActiveDeadlineSeconds(int64(900)).Return(tf),
				tf.EXPECT().SetDeadlineCleaning(5*time.Minute).Return(tf),
				tf.EXPECT().SetDeadlinePod(5*time.Minute).Return(tf),
				tf.EXPECT().SetDeadlineJob(15*time.Minute).Return(tf),
			)

			actual, err := NewTerraformer(factory, &config, &credentials, purpose, namespace, name)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(BeIdenticalTo(tf))
		})
	})
})
