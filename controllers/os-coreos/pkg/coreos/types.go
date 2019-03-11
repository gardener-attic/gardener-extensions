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

package coreos

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// CloudConfig is a structure containing the relevant fields for generating the Config
// Container Linux specific cloud config. It can be marshalled to YAML.
type CloudConfig struct {
	// Config contains CoreOS specific configuration.
	CoreOS Config `yaml:"coreos,omitempty"`
	// WriteFiles is a list of files that will be written onto the disk of the machine where
	// the cloud config is executed on.
	WriteFiles []File `yaml:"write_files,omitempty"`
}

// Config contains CoreOS specific configuration.
type Config struct {
	// Update contains configuration for the Container Linux update procedure.
	Update Update `yaml:"update,omitempty"`
	// Units is a list of units that are translated to systemd later.
	Units []Unit `yaml:"units,omitempty"`
}

// Update contains configuration for the Container Linux update procedure.
type Update struct {
	// RebootStrategy defines the strategy for reboots.
	RebootStrategy string `yaml:"reboot_strategy,omitempty"`
	// Group describes the update group.
	Group string `yaml:"group,omitempty"`
	// Server describes the update server.
	Server string `yaml:"server,omitempty"`
}

// Unit gets translated to a systemd unit.
type Unit struct {
	// Name is the name of the unit.
	Name string `yaml:"name,omitempty"`
	// Mask defines whether unit is masked or not.
	Mask bool `yaml:"mask,omitempty"`
	// Enable defines whether unit is enabled or not.
	Enable bool `yaml:"enable,omitempty"`
	// Runtime defines the runtime of the unit.
	Runtime bool `yaml:"runtime,omitempty"`
	// Content defines the actual systemd specific content of the unit.
	Content string `yaml:"content,omitempty"`
	// Command describes the command of the unit.
	Command string `yaml:"command,omitempty"`
	// DropIns is a list of drop-in units.
	DropIns []UnitDropIn `yaml:"drop_ins,omitempty"`
}

// UnitDropIn is a structure for information about a drop-in unit.
type UnitDropIn struct {
	// Name is the name of the drop-in.
	Name string `yaml:"name,omitempty"`
	// Content is the content of the drop-in.
	Content string `yaml:"content,omitempty"`
}

// File is a file that gets written to onto the disk of the machine where the cloud config is
// executed on.
type File struct {
	// Encoding is the encoding of the file, e.g. base64.
	Encoding string `yaml:"encoding,omitempty"`
	// Content is the actual content of the file.
	Content string `yaml:"content,omitempty"`
	// Owner describes the owner of the file.
	Owner string `yaml:"owner,omitempty"`
	// Path is the path on the disk where the file should get written to.
	Path string `yaml:"path,omitempty"`
	// RawFilePermissions describes the permissions for the file, e.g. 0777.
	RawFilePermissions string `yaml:"permissions,omitempty"`
}

// String returns the string representation of the CloudConfig structure.
func (c CloudConfig) String() (string, error) {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("#cloud-config\n\n%s", string(bytes)), nil
}
