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

const (
	hostLocal  = "host-local"
	usePodCIDR = "usePodCidr"
)

// ComputeCalicoChartValues computes the values for the calico chart.
func ComputeCalicoChartValues(network *extensionsv1alpha1.Network, config *calicov1alpha1.NetworkConfig) map[string]interface{} {
	var (
		calicoChartValues = map[string]interface{}{
			"images": map[string]interface{}{
				calico.CNIImageName:                                        imagevector.CalicoCNIImage(),
				calico.TyphaImageName:                                      imagevector.CalicoTyphaImage(),
				calico.KubeControllersImageName:                            imagevector.CalicoKubeControllersImage(),
				calico.NodeImageName:                                       imagevector.CalicoNodeImage(),
				calico.PodToDaemonFlexVolumeDriverImageName:                imagevector.CalicoFlexVolumeDriverImage(),
				calico.TyphaClusterProportionalAutoscalerImageName:         imagevector.TyphaClusterProportionalAutoscalerImage(),
				calico.TyphaClusterProportionalVerticalAutoscalerImageName: imagevector.TyphaClusterProportionalVerticalAutoscalerImage(),
			},
			"global": map[string]string{
				"podCIDR": network.Spec.PodCIDR,
			},
			"ipip": calicov1alpha1.Always,
		}
		calicoConfigValues = map[string]interface{}{
			"backend": calicov1alpha1.Bird,
		}
		ipamConfig = map[string]interface{}{
			"type":   hostLocal,
			"subnet": usePodCIDR,
		}
	)

	if config != nil {
		switch config.Backend {
		case calicov1alpha1.Bird, calicov1alpha1.None:
			calicoConfigValues["backend"] = config.Backend
		}

		if config.IPAM != nil {
			if len(config.IPAM.Type) > 0 {
				ipamConfig["type"] = config.IPAM.Type
			}
			if config.IPAM.Type == hostLocal && config.IPAM.CIDR != nil {
				ipamConfig["subnet"] = *config.IPAM.CIDR
			}
		}

		if config.IPAutoDetectionMethod != nil {
			calicoChartValues["ipAutodetectionMethod"] = *config.IPAutoDetectionMethod
		}
		if config.IPIP != nil {
			switch *config.IPIP {
			case calicov1alpha1.Always, calicov1alpha1.Never, calicov1alpha1.Off, calicov1alpha1.CrossSubnet:
				calicoChartValues["ipip"] = *config.IPIP
			}
		}
	}

	calicoConfigValues["ipam"] = ipamConfig
	calicoChartValues["config"] = calicoConfigValues

	return calicoChartValues
}
