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

package customizer

import (
	customizer "github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig/oscommon/customizer"
	generator "github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type JeosCustomizer struct {
}

func NewCustomizer() (customizer.Customizer, error) {
	return &JeosCustomizer{}, nil
}

func (c *JeosCustomizer) Customize(osconf *generator.OperatingSystemConfig, extension *runtime.RawExtension) (*generator.OperatingSystemConfig, error) {

	customization := JeosOperatingSystemCustomization{}
	err := yaml.Unmarshal(extension.Raw, &customization)
	if err != nil {
		return nil, err
	}

	var units []*generator.Unit
	var files []*generator.File
	var commands []string

	var customized = &generator.OperatingSystemConfig{}

	for _, service := range customization.EnableServices {
		cmd := "systemctl enable " + service
		commands = append(commands, cmd)
	}

	for _, cmd := range customization.BootCommands {
		commands = append(commands, cmd)
	}

	customized.Files = append(osconf.Files, files...)
	customized.Units = append(osconf.Units, units...)
	customized.Commands = append(osconf.Commands, commands...)
	customized.Bootstrap = osconf.Bootstrap
	customized.Path = osconf.Path

	return customized, nil
}

type JeosOperatingSystemCustomization struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// systmctl services to enable
	EnableServices []string

	// commands to execute at boot time
	BootCommands []string
}
