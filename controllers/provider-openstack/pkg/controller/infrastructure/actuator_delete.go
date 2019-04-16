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

package infrastructure

import (
	"context"
	"fmt"
	"github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/openstack"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

func (a *actuator) delete(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	tf, err := a.newTerraformer(openstack.TerrformerPurposeInfra, infrastructure.Namespace, infrastructure.Name)
	if err != nil {
		return fmt.Errorf("could not create the Terraformer: %+v", err)
	}

	creds, err := getCredentials(ctx, a.client, infrastructure)
	if err != nil {
		return err
	}

	return tf.
		SetVariablesEnvironment(generateTerraformInfraVariablesEnvironment(creds)).
		Destroy()
}
