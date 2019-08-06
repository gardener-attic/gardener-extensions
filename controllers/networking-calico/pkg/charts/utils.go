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

package charts

import (
	calicov1alpha1 "github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/calico"
	"github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/imagevector"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

// ComputeCalicoChartValues computes the values for the calico chart.
func ComputeCalicoChartValues(network *extensionsv1alpha1.Network, config *calicov1alpha1.NetworkConfig) map[string]interface{} {
	return map[string]interface{}{
		"images": map[string]interface{}{
			calico.CNIImageName:             imagevector.CalicoCNIImage(),
			calico.TyphaImageName:           imagevector.CalicoTyphaImage(),
			calico.KubeControllersImageName: imagevector.CalicoKubeControllersImage(),
			calico.NodeImageName:            imagevector.CalicoNodeImage(),
		},
		"config": map[string]string{
			"backend": string(config.Backend),
		},
		"global": map[string]string{
			"podCIDR": network.Spec.PodCIDR,
		},
	}
}
