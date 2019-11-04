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
	"path/filepath"
	"time"

	packetv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/apis/packet/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/packet"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	controllererrors "github.com/gardener/gardener-extensions/pkg/controller/error"
	"github.com/gardener/gardener-extensions/pkg/terraformer"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
)

func (a *actuator) reconcile(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	providerSecret := &corev1.Secret{}
	if err := a.client.Get(ctx, kutil.Key(infrastructure.Spec.SecretRef.Namespace, infrastructure.Spec.SecretRef.Name), providerSecret); err != nil {
		return err
	}

	terraformConfig := GenerateTerraformInfraConfig(infrastructure, string(providerSecret.Data[packet.ProjectID]))

	chartRenderer, err := chartrenderer.NewForConfig(a.restConfig)
	if err != nil {
		return fmt.Errorf("could not create chart renderer: %+v", err)
	}

	release, err := chartRenderer.Render(filepath.Join(packet.InternalChartsPath, "packet-infra"), "packet-infra", infrastructure.Namespace, terraformConfig)
	if err != nil {
		return fmt.Errorf("could not render Terraform chart: %+v", err)
	}

	tf, err := a.newTerraformer(packet.TerraformerPurposeInfra, infrastructure.Namespace, infrastructure.Name)
	if err != nil {
		return fmt.Errorf("could not create terraformer object: %+v", err)
	}

	if err := tf.
		SetVariablesEnvironment(generateTerraformInfraVariablesEnvironment(providerSecret)).
		InitializeWith(terraformer.DefaultInitializer(
			a.client,
			release.FileContent("main.tf"),
			release.FileContent("variables.tf"),
			[]byte(release.FileContent("terraform.tfvars"))),
		).
		Apply(); err != nil {

		a.logger.Error(err, "failed to apply the terraform config", "infrastructure", infrastructure.Name)
		return &controllererrors.RequeueAfterError{
			Cause:        err,
			RequeueAfter: 30 * time.Second,
		}
	}

	return a.updateProviderStatus(ctx, tf, infrastructure)
}

// GenerateTerraformInfraConfig generates the Packet Terraform configuration based on the given infrastructure and project.
func GenerateTerraformInfraConfig(infrastructure *extensionsv1alpha1.Infrastructure, projectID string) map[string]interface{} {
	return map[string]interface{}{
		"packet": map[string]interface{}{
			"projectID": projectID,
		},
		"sshPublicKey": string(infrastructure.Spec.SSHPublicKey),
		"clusterName":  infrastructure.Namespace,
		"outputKeys": map[string]interface{}{
			"sshKeyID": packet.SSHKeyID,
		},
	}
}

func (a *actuator) updateProviderStatus(ctx context.Context, tf *terraformer.Terraformer, infrastructure *extensionsv1alpha1.Infrastructure) error {
	outputVarKeys := []string{
		packet.SSHKeyID,
	}

	output, err := tf.GetStateOutputVariables(outputVarKeys...)
	if err != nil {
		return err
	}

	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.client, infrastructure, func() error {
		infrastructure.Status.ProviderStatus = &runtime.RawExtension{
			Object: &packetv1alpha1.InfrastructureStatus{
				TypeMeta: metav1.TypeMeta{
					APIVersion: packetv1alpha1.SchemeGroupVersion.String(),
					Kind:       "InfrastructureStatus",
				},
				SSHKeyID: output[packet.SSHKeyID],
			},
		}
		return nil
	})
}
