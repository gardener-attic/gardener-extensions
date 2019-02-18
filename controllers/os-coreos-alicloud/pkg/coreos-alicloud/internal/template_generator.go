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

//go:generate packr2

package internal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"k8s.io/apimachinery/pkg/util/runtime"
	"path"
	"text/template"
)

var cloudInitTemplate *template.Template

// DefaultUnitsPath is the default CoreOS path where to store units at.
const DefaultUnitsPath = "/etc/systemd/system"

func init() {
	box := packr.New("templates", "./templates")

	cloudInitTemplateString, err := box.FindString("cloud-init.sh.template")
	runtime.Must(err)

	cloudInitTemplate, err = template.New("cloud-init.sh").Parse(cloudInitTemplateString)
	runtime.Must(err)
}

type fileData struct {
	Path        string
	Content     string
	Dirname     string
	Permissions *string
}

type unitData struct {
	Path    string
	Name    string
	Content *string
	DropIns *dropInsData
}

type dropInsData struct {
	Path  string
	Items []*dropInData
}

type dropInData struct {
	Path    string
	Content string
}

type initScriptData struct {
	Files     []*fileData
	Units     []*unitData
	Bootstrap bool
}

// CloudInitGenerator generates cloud-init scripts.
type CloudInitGenerator struct {
	unitsPath string
}

func b64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Generate generates a cloud-init script from the given OperatingSystemConfig.
func (t *CloudInitGenerator) Generate(data *OperatingSystemConfig) ([]byte, error) {
	var tFiles []*fileData
	for _, file := range data.Files {
		tFile := &fileData{
			Path:    file.Path,
			Content: b64(file.Content),
			Dirname: path.Dir(file.Path),
		}
		if file.Permissions != nil {
			permissions := fmt.Sprintf("%04o", *file.Permissions)
			tFile.Permissions = &permissions
		}
		tFiles = append(tFiles, tFile)
	}

	var tUnits []*unitData
	for _, unit := range data.Units {
		var content *string
		if unit.Content != nil {
			encoded := b64(unit.Content)
			content = &encoded
		}
		tUnit := &unitData{
			Name:    unit.Name,
			Path:    path.Join(t.unitsPath, unit.Name),
			Content: content,
		}
		if len(unit.DropIns) != 0 {
			dropInPath := path.Join(t.unitsPath, fmt.Sprintf("%s.d", unit.Name))

			var items []*dropInData
			for _, dropIn := range unit.DropIns {
				items = append(items, &dropInData{
					Path:    path.Join(dropInPath, dropIn.Name),
					Content: b64(dropIn.Content),
				})
			}
			tUnit.DropIns = &dropInsData{
				Path:  dropInPath,
				Items: items,
			}
		}

		tUnits = append(tUnits, tUnit)
	}

	var buf bytes.Buffer
	if err := cloudInitTemplate.Execute(&buf, &initScriptData{
		Files:     tFiles,
		Units:     tUnits,
		Bootstrap: data.Bootstrap,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// NewCloudInitGenerator creates a new CloudInitGenerator with the given units path.
func NewCloudInitGenerator(unitsPath string) *CloudInitGenerator {
	return &CloudInitGenerator{unitsPath}
}
