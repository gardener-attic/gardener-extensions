// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package helper

import (
	"fmt"

	api "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/config"
)

// FindImageForRegion takes a list of machine images, and the desired image name, version and region. It tries
// to find the image with the given name, version and region. If it cannot be found then an error
// is returned.
func FindImageForRegion(profileImages []api.MachineImages, configImages []config.MachineImage, imageName, imageVersion, regionID string) (string, error) {
	for _, machineImage := range profileImages {
		if machineImage.Name == imageName {
			for _, version := range machineImage.Versions {
				if imageVersion == version.Version {
					return version.ID, nil
				}
			}
		}
	}

	for _, machineImage := range configImages {
		if machineImage.Name != imageName || machineImage.Version != imageVersion {
			continue
		}

		for _, region := range machineImage.Regions {
			if region.Region == regionID {
				return region.ImageID, nil
			}
		}
	}

	return "", fmt.Errorf("could not find an image for region %q and name %q in version %q", regionID, imageName, imageVersion)
}
