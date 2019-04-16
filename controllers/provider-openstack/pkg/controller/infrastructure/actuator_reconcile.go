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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"path/filepath"
	"time"

	openstackv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/apis/openstack/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/provider-openstack/pkg/openstack"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	controllererrors "github.com/gardener/gardener-extensions/pkg/controller/error"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/operation/terraformer"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultRouterID = "${openstack_networking_router_v2.router.id}"
	// DomainName is a constant for the key in a cloud provider secret that holds the OpenStack domain name.
	DomainName = "domainName"
	// TenantName is a constant for the key in a cloud provider secret that holds the OpenStack tenant name.
	TenantName = "tenantName"
	// UserName is a constant for the key in a cloud provider secret and backup secret that holds the OpenStack username.
	UserName = "username"
	// Password is a constant for the key in a cloud provider secret and backup secret that holds the OpenStack password.
	Password = "password"
)

type credentials struct {
	DomainName string
	TenantName string
}

func (a *actuator) reconcile(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	infrastructureConfig := &openstackv1alpha1.InfrastructureConfig{}
	if _, _, err := a.decoder.Decode(infrastructure.Spec.ProviderConfig.Raw, nil, infrastructureConfig); err != nil {
		return fmt.Errorf("could not decode provider config: %+v", err)
	}

	providerSecret := &corev1.Secret{}
	if err := a.client.Get(ctx, kutil.Key(infrastructure.Spec.SecretRef.Namespace, infrastructure.Spec.SecretRef.Name), providerSecret); err != nil {
		return err
	}

	terraformConfig, err := generateTerraformInfraConfig(ctx, infrastructure, infrastructureConfig, cluster, extractCredentials(providerSecret))
	if err != nil {
		return fmt.Errorf("failed to generate Terraform config: %+v", err)
	}

	chartRenderer, err := chartrenderer.NewForConfig(a.restConfig)
	if err != nil {
		return fmt.Errorf("could not create chart renderer: %+v", err)
	}

	release, err := chartRenderer.Render(filepath.Join(openstack.InternalChartsPath, "openstack-infra"), "openstack-infra", infrastructure.Namespace, terraformConfig)
	if err != nil {
		return fmt.Errorf("could not render Terraform chart: %+v", err)
	}

	tf, err := a.newTerraformer(openstack.TerrformerPurposeInfra, infrastructure.Namespace, infrastructure.Name)
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

		return &controllererrors.RequeueAfterError{
			Cause:        err,
			RequeueAfter: 30 * time.Second,
		}
	}

	if err := a.updateProviderStatus(ctx, tf, infrastructure, infrastructureConfig); err != nil {
		return fmt.Errorf("failed to update the provider status in the Infrastructure resource: %+v", err)
	}
	return nil
}

func (a *actuator) updateProviderStatus(ctx context.Context, tf *terraformer.Terraformer, infrastructure *extensionsv1alpha1.Infrastructure, infrastructureConfig *openstackv1alpha1.InfrastructureConfig) error {
	outputVarKeys := []string{
		openstack.SSHKeyName,
		openstack.RouterID,
		openstack.NetworkID,
		openstack.SubnetID,
		openstack.FloatingNetworkID,
		openstack.SecurityGroupID,
		openstack.SecurityGroupName,
	}

	output, err := tf.GetStateOutputVariables(outputVarKeys...)
	if err != nil {
		return err
	}

	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.client, infrastructure, func() error {
		infrastructure.Status.ProviderStatus = &runtime.RawExtension{
			Object: &openstackv1alpha1.InfrastructureStatus{
				TypeMeta: metav1.TypeMeta{
					APIVersion: openstackv1alpha1.SchemeGroupVersion.String(),
					Kind:       "InfrastructureStatus",
				},
				Router: openstackv1alpha1.RouterStatus{
					ID: output[openstack.RouterID],
				},
				Network: openstackv1alpha1.NetworkStatus{
					ID: output[openstack.NetworkID],
					Subnets: []openstackv1alpha1.Subnet{
						{
							ID:      output[openstack.SubnetID],
							Purpose: openstackv1alpha1.PurposeNodes,
						},
					},
				},
				Node: openstackv1alpha1.NodeStatus{
					KeyName: output[openstack.SSHKeyName],
				},
			},
		}
		return nil
	})
}

func extractCredentials(providerSecret *corev1.Secret) *credentials {
	return &credentials{
		DomainName: string(providerSecret.Data[DomainName]),
		TenantName: string(providerSecret.Data[TenantName]),
	}
}

func generateTerraformInfraConfig(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure, infrastructureConfig *openstackv1alpha1.InfrastructureConfig, cluster *extensionscontroller.Cluster, credentials *credentials) (map[string]interface{}, error) {
	var (
		routerID     = defaultRouterID
		createRouter = true
	)
	if router := infrastructureConfig.Networks.Router; router != nil {
		createRouter = false
		routerID = router.ID
	}
	return map[string]interface{}{
		"openstack": map[string]interface{}{
			"authURL":          cluster.CloudProfile.Spec.OpenStack.KeyStoneURL,
			"domainName":       credentials.DomainName,
			"tenantName":       credentials.TenantName,
			"region":           infrastructure.Spec.Region,
			"floatingPoolName": infrastructureConfig.FloatingPoolName,
		},
		"create": map[string]interface{}{
			"router": createRouter,
		},
		"dnsServers":   cluster.CloudProfile.Spec.OpenStack.DNSServers,
		"sshPublicKey": infrastructure.Spec.SSHPublicKey,
		"router": map[string]interface{}{
			"id": routerID,
		},
		"clusterName": infrastructure.Namespace,
		"networks": map[string]interface{}{
			"worker": infrastructureConfig.Networks.Worker,
		},
	}, nil
}
