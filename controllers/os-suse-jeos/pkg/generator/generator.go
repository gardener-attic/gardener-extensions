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

package generator

import (
	template_gen "github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig/oscommon/template"
	"github.com/gobuffalo/packr/v2"
	"text/template"
)

var cmd = "/usr/bin/cloud-init clean && /usr/bin/cloud-init --file %s init"

//go:generate packr2

// NewCloudInitGenerator creates a new Generator using the template file for suse-jeos
func NewCloudInitGenerator() (*template_gen.CloudInitGenerator, error) {
	box := packr.New("templates", "./templates")
	cloudInitTemplateString, err := box.FindString("cloud-init.template")
	if err != nil {
		return nil, err
	}

	cloudInitTemplate, err := template.New("cloud-init").Parse(cloudInitTemplateString)
	if err != nil {
		return nil, err
	}
	generator := template_gen.NewCloudInitGenerator(cloudInitTemplate, template_gen.DefaultUnitsPath, cmd)
	return generator, nil
}
