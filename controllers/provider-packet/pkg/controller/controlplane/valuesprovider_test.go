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

package controlplane

import (
	"context"
	"encoding/json"

	apispacket "github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/apis/packet"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/operation/common"

	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	namespace = "test"
)

var _ = Describe("ValuesProvider", func() {
	var (
		ctrl *gomock.Controller

		// Build scheme
		scheme = runtime.NewScheme()
		_      = apispacket.AddToScheme(scheme)

		cp = &extensionsv1alpha1.ControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "control-plane",
				Namespace: namespace,
			},
			Spec: extensionsv1alpha1.ControlPlaneSpec{
				ProviderConfig: &runtime.RawExtension{
					Raw: encode(&apispacket.ControlPlaneConfig{}),
				},
				InfrastructureProviderStatus: &runtime.RawExtension{
					Raw: encode(&apispacket.InfrastructureStatus{}),
				},
			},
		}

		cluster = &extensionscontroller.Cluster{
			Shoot: &gardenv1beta1.Shoot{},
		}

		checksums = map[string]string{
			common.CloudProviderSecretName: "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			"cloud-controller-manager":     "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
		}

		ccmChartValues = map[string]interface{}{
			"replicas": 1,
			"podAnnotations": map[string]interface{}{
				"checksum/secret-cloud-controller-manager": "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
				"checksum/secret-cloudprovider":            "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			},
		}

		logger = log.Log.WithName("test")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#GetConfigChartValues", func() {
		It("should return correct config chart values", func() {
			// Create valuesProvider
			vp := NewValuesProvider(logger)
			err := vp.(inject.Scheme).InjectScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			// Call GetConfigChartValues method and check the result
			values, err := vp.GetConfigChartValues(context.TODO(), cp, cluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(BeNil())
		})
	})

	Describe("#GetControlPlaneChartValues", func() {
		It("should return correct control plane chart values", func() {
			// Create valuesProvider
			vp := NewValuesProvider(logger)
			err := vp.(inject.Scheme).InjectScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			// Call GetControlPlaneChartValues method and check the result
			values, err := vp.GetControlPlaneChartValues(context.TODO(), cp, cluster, checksums)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(ccmChartValues))
		})
	})
})

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}
