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

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	awsclient "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws/client"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	controllererrors "github.com/gardener/gardener-extensions/pkg/controller/error"

	gardencorev1alpha1helper "github.com/gardener/gardener/pkg/apis/core/v1alpha1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	glogger "github.com/gardener/gardener/pkg/logger"
	"github.com/gardener/gardener/pkg/operation/terraformer"
	"github.com/gardener/gardener/pkg/utils/flow"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (a *actuator) delete(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	tf, err := a.newTerraformer(aws.TerrformerPurposeInfra, infrastructure.Namespace, infrastructure.Name)
	if err != nil {
		return fmt.Errorf("could not create the Terraformer: %+v", err)
	}

	configExists, err := tf.ConfigExists()
	if err != nil {
		return fmt.Errorf("terraform configuration was not found: %+v", err)
	}

	stateVariables, err := tf.GetStateOutputVariables(aws.VPCIDKey)
	if err != nil {
		if apierrors.IsNotFound(err) || terraformer.IsVariablesNotFoundError(err) {
			a.logger.Info("Skipping explicit AWS load balancer and security group deletion because not all variables have been found in the Terraform state.")
			return nil
		}
		return err
	}
	vpcID := stateVariables[aws.VPCIDKey]

	providerSecret := &corev1.Secret{}
	if err := a.client.Get(ctx, kutil.Key(infrastructure.Spec.SecretRef.Namespace, infrastructure.Spec.SecretRef.Name), providerSecret); err != nil {
		return err
	}

	awsClient, err := awsclient.NewClient(string(providerSecret.Data[aws.AccessKeyID]), string(providerSecret.Data[aws.SecretAccessKey]), infrastructure.Spec.Region)
	if err != nil {
		return err
	}

	var (
		g = flow.NewGraph("AWS infrastructure destruction")

		destroyKubernetesLoadBalancersAndSecurityGroups = g.Add(flow.Task{
			Name: "Destroying Kubernetes load balancers and security groups",
			Fn: flow.TaskFn(func(ctx context.Context) error {
				if err := a.destroyKubernetesLoadBalancersAndSecurityGroups(ctx, awsClient, vpcID, infrastructure.Namespace); err != nil {
					return gardencorev1alpha1helper.DetermineError(fmt.Sprintf("Failed to destroy load balancers and security groups: %+v", err.Error()))
				}
				return nil
			}).RetryUntilTimeout(10*time.Second, 5*time.Minute).DoIf(configExists),
		})

		_ = g.Add(flow.Task{
			Name:         "Destroying Shoot infrastructure",
			Fn:           flow.SimpleTaskFn(tf.SetVariablesEnvironment(generateTerraformInfraVariablesEnvironment(providerSecret)).Destroy),
			Dependencies: flow.NewTaskIDs(destroyKubernetesLoadBalancersAndSecurityGroups),
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

func (a *actuator) destroyKubernetesLoadBalancersAndSecurityGroups(ctx context.Context, awsClient awsclient.Interface, vpcID, clusterName string) error {
	loadBalancers, err := awsClient.ListKubernetesELBs(ctx, vpcID, clusterName)
	if err != nil {
		return err
	}
	securityGroups, err := awsClient.ListKubernetesSecurityGroups(ctx, vpcID, clusterName)
	if err != nil {
		return err
	}

	for _, loadBalancerName := range loadBalancers {
		if err := awsClient.DeleteELB(ctx, loadBalancerName); err != nil {
			return err
		}
	}
	for _, securityGroupID := range securityGroups {
		if err := awsClient.DeleteSecurityGroup(ctx, securityGroupID); err != nil {
			return err
		}
	}

	return nil
}
