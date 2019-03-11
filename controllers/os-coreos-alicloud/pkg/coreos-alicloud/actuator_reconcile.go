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
	"context"
	"fmt"

	"github.com/gardener/gardener-extensions/controllers/os-coreos-alicloud/pkg/coreos-alicloud/internal"
	"github.com/gardener/gardener-extensions/controllers/os-coreos-alicloud/pkg/coreos-alicloud/internal/cloudinit"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (a *actuator) reconcile(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) ([]byte, *string, []string, error) {
	cloudConfig, err := a.cloudConfigFromOperatingSystemConfig(ctx, config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Could not generate cloud config: %v", err)
	}

	var command *string
	if path := config.Spec.ReloadConfigFilePath; path != nil {
		cmd := fmt.Sprintf("/usr/bin/env bash %s", *path)
		command = &cmd
	}

	return []byte(cloudConfig), command, operatingSystemConfigUnitNames(config), nil
}

func (a *actuator) cloudConfigFromOperatingSystemConfig(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) ([]byte, error) {
	files := make([]*internal.File, 0, len(config.Spec.Files))
	for _, file := range config.Spec.Files {
		data, err := a.dataForFileContent(ctx, config.Namespace, &file.Content)
		if err != nil {
			return nil, err
		}

		files = append(files, &internal.File{Path: file.Path, Content: data, Permissions: file.Permissions})
	}

	units := make([]*internal.Unit, 0, len(config.Spec.Units))
	for _, unit := range config.Spec.Units {
		var content []byte
		if unit.Content != nil {
			content = []byte(*unit.Content)
		}

		dropIns := make([]*internal.DropIn, 0, len(unit.DropIns))
		for _, dropIn := range unit.DropIns {
			dropIns = append(dropIns, &internal.DropIn{Name: dropIn.Name, Content: []byte(dropIn.Content)})
		}
		units = append(units, &internal.Unit{Name: unit.Name, Content: content, DropIns: dropIns})
	}

	return internal.NewCloudInitGenerator(internal.DefaultUnitsPath).
		Generate(&internal.OperatingSystemConfig{
			Bootstrap: config.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeProvision,
			Files:     files,
			Units:     units,
		})
}

func (a *actuator) dataForFileContent(ctx context.Context, namespace string, content *extensionsv1alpha1.FileContent) ([]byte, error) {
	if inline := content.Inline; inline != nil {
		if len(inline.Encoding) == 0 {
			return []byte(inline.Data), nil
		}
		return cloudinit.Decode(inline.Encoding, []byte(inline.Data))
	}

	key := client.ObjectKey{Namespace: namespace, Name: content.SecretRef.Name}
	secret := &corev1.Secret{}
	if err := a.client.Get(ctx, key, secret); err != nil {
		return nil, err
	}
	return secret.Data[content.SecretRef.DataKey], nil
}

func operatingSystemConfigUnitNames(config *extensionsv1alpha1.OperatingSystemConfig) []string {
	unitNames := make([]string, 0, len(config.Spec.Units))
	for _, unit := range config.Spec.Units {
		unitNames = append(unitNames, unit.Name)
	}
	return unitNames
}
