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
	"time"

	"github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/openstack"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	controllererrors "github.com/gardener/gardener-extensions/pkg/controller/error"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	glogger "github.com/gardener/gardener/pkg/logger"
	"github.com/gardener/gardener/pkg/utils/flow"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

func (a *actuator) delete(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	tf, err := a.newTerraformer(openstack.TerrformerPurposeInfra, infrastructure.Namespace, infrastructure.Name)
	if err != nil {
		return fmt.Errorf("could not create the Terraformer: %+v", err)
	}

	providerSecret := &corev1.Secret{}
	if err := a.client.Get(ctx, kutil.Key(infrastructure.Spec.SecretRef.Namespace, infrastructure.Spec.SecretRef.Name), providerSecret); err != nil {
		return err
	}

	var (
		g = flow.NewGraph("OpenStack infrastructure destruction")

		_ = g.Add(flow.Task{
			Name: "Destroying Shoot infrastructure",
			Fn:   flow.SimpleTaskFn(tf.SetVariablesEnvironment(generateTerraformInfraVariablesEnvironment(providerSecret)).Destroy),
		})

		f = g.Compile()
	)

	if err := f.Run(flow.Opts{Context: ctx, Logger: glogger.NewFieldLogger(glogger.NewLogger("info"), "infrastructure", infrastructure.Name)}); err != nil {
		return &controllererrors.RequeueAfterError{
			Cause:        flow.Causes(err),
			RequeueAfter: 30 * time.Second,
		}
	}

	return nil
}
