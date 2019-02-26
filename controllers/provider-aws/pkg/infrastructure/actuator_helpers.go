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
	"strconv"
	"strings"

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/awsclient"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws/v1alpha1"
	"github.com/gardener/gardener-extensions/pkg/controller"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/logger"
	"github.com/gardener/gardener/pkg/operation/terraformer"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (a *actuator) updateStatusError(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure, lastOperationType extensionsv1alpha1.LastOperationType, description string, progress int, err error) {
	infrastructure.Status.ObservedGeneration = infrastructure.Generation
	infrastructure.Status.LastOperation, infrastructure.Status.LastError = controller.ReconcileError(lastOperationType, fmt.Sprintf("%s: %v", description, err), progress)
	if err := a.client.Status().Update(ctx, infrastructure); err != nil {
		a.logger.Error(err, "Could not update infrastructure infrastructure status after update error", "infrastructure", infrastructure.Name)
	}
}

func (a *actuator) updateStatusSuccess(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure, lastOperationType extensionsv1alpha1.LastOperationType, description string) error {
	infrastructure.Status.ObservedGeneration = infrastructure.Generation
	infrastructure.Status.LastOperation, infrastructure.Status.LastError = controller.ReconcileSucceeded(lastOperationType, description)
	return a.client.Status().Update(ctx, infrastructure)
}

func (a *actuator) getTerraformer(purpose, namespace, name string) (*terraformer.Terraformer, error) {
	return terraformer.NewForConfig(logger.NewLogger("info"), a.config, purpose, namespace, name, a.terraformerImage)
}

func (a *actuator) destroyKubernetesLoadBalancersAndSecurityGroups(ctx context.Context, vpcIDKey, namespace string, tf *terraformer.Terraformer, awsClient awsclient.ClientInterface) error {
	if _, err := tf.GetState(); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	stateVariables, err := tf.GetStateOutputVariables(vpcIDKey)
	if err != nil {
		if terraformer.IsVariablesNotFoundError(err) {
			a.logger.Info("Skipping explicit AWS load balancer and security group deletion because not all variables have been found in the Terraform state.")
			return nil
		}
		return err
	}
	vpcID := stateVariables[vpcIDKey]
	// Find load balancers and security groups.
	loadBalancers, err := awsClient.ListKubernetesELBs(ctx, vpcID, namespace)
	if err != nil {
		return err
	}
	securityGroups, err := awsClient.ListKubernetesSecurityGroups(ctx, vpcID, namespace)
	if err != nil {
		return err
	}

	// Destroy load balancers and security groups.
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

func (a *actuator) injectProviderStateIntoStatus(ctx context.Context, tf *terraformer.Terraformer, infrastructure *extensionsv1alpha1.Infrastructure) error {
	outputVarKeys := []string{
		aws.VPCIDKey,
		aws.SSHKeyName,
		aws.IAMInstanceProfileNodes,
		aws.NodesRole,
		aws.SecurityGroupsNodes,
	}

	for zoneIndex := range infrastructure.Spec.Zones {
		outputVarKeys = append(outputVarKeys, fmt.Sprintf("%s%d", aws.SubnetNodesPrefix, zoneIndex))
		outputVarKeys = append(outputVarKeys, fmt.Sprintf("%s%d", aws.SubnetPublicPrefix, zoneIndex))
	}

	values, err := tf.GetStateOutputVariables(outputVarKeys...)
	if err != nil {
		return err
	}

	subnets, err := getSubnets(infrastructure, values)
	if err != nil {
		return err
	}

	var (
		purpose          = v1alpha1.PurposeNodes
		instanceProfiles = []v1alpha1.InstanceProfile{
			{
				Purpose: &purpose,
				Name:    values[aws.IAMInstanceProfileNodes],
			},
		}
		roles = []v1alpha1.Role{
			{
				Purpose: &purpose,
				ARN:     values[aws.NodesRole],
			},
		}
		securityGroups = []v1alpha1.SecurityGroup{
			{
				Name: aws.SecurityGroupsNodes,
				ID:   values[aws.SecurityGroupsNodes],
			},
		}
	)

	infrastructure.Status.ProviderStatus = &runtime.RawExtension{
		Object: &v1alpha1.InfrastructureStatus{
			VPC: v1alpha1.VPCStatus{
				ID:             values[aws.VPCIDKey],
				Subnets:        subnets,
				SecurityGroups: securityGroups,
			},
			EC2: v1alpha1.EC2{
				KeyName: values[aws.SSHKeyName],
			},
			IAM: v1alpha1.IAM{
				InstanceProfiles: instanceProfiles,
				Roles:            roles,
			},
		},
	}
	return a.client.Status().Update(ctx, infrastructure)
}

// generateTerraformInfraConfigValues creates the Terraform variables and the Terraform config (for the infrastructure)
// and returns them (these values will be stored as a ConfigMap and a Secret in the Garden cluster.
func generateTerraformInfraConfigValues(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure, infrastructureConfig *v1alpha1.InfrastructureConfig, providerSecret *corev1.Secret) (map[string]interface{}, error) {
	var (
		dhcpDomainName    = "ec2.internal"
		createVPC         = true
		vpcID             = "${aws_vpc.vpc.id}"
		internetGatewayID = "${aws_internet_gateway.igw.id}"
		vpcCIDR           = infrastructureConfig.Networks.VPC.CIDR
		region            = infrastructure.Spec.Region
	)

	if infrastructure.Spec.Region != "us-east-1" {
		dhcpDomainName = fmt.Sprintf("%s.compute.internal", infrastructure.Spec.Region)
	}

	awsClient, err := awsclient.NewClient(string(providerSecret.Data[aws.AccessKeyID]), string(providerSecret.Data[aws.SecretAccessKey]), region)
	if err != nil {
		return nil, err
	}
	// check if we should use an existing VPC or create a new one
	if infrastructureConfig.Networks.VPC.ID != nil {
		createVPC = false
		vpcID = *infrastructureConfig.Networks.VPC.ID
		igwID, err := awsClient.GetInternetGateway(ctx, vpcID)
		if err != nil {
			return nil, err
		}
		internetGatewayID = igwID
	} else if infrastructureConfig.Networks.VPC.CIDR != nil {
		vpcCIDR = infrastructureConfig.Networks.VPC.CIDR
	}

	return map[string]interface{}{
		"aws": map[string]interface{}{
			"region": infrastructure.Spec.Region,
		},
		"create": map[string]interface{}{
			"vpc": createVPC,
		},
		"sshPublicKey": string(infrastructure.Spec.SSHPublicKey),
		"vpc": map[string]interface{}{
			"id":                vpcID,
			"cidr":              vpcCIDR,
			"dhcpDomainName":    dhcpDomainName,
			"internetGatewayID": internetGatewayID,
		},
		"clusterName": infrastructure.Namespace,
		"zones":       getZones(infrastructureConfig),
		"outputKeys": map[string]interface{}{
			"vpcIdKey":                   aws.VPCIDKey,
			"subnetsPublicPrefix":        aws.SubnetPublicPrefix,
			"subnetsNodesPrefix":         aws.SubnetNodesPrefix,
			"securityGroupsNodes":        aws.SecurityGroupsNodes,
			"sshKeyName":                 aws.SSHKeyName,
			"iamInstanceProfileNodes":    aws.IAMInstanceProfileNodes,
			"iamInstanceProfileBastions": aws.IAMInstanceProfileBastions,
			"nodesRole":                  aws.NodesRole,
			"bastionsRole":               aws.BastionsRole,
		},
	}, nil
}

func getZones(infraProviderConfig *v1alpha1.InfrastructureConfig) []map[string]interface{} {
	var zones []map[string]interface{}
	for _, zone := range infraProviderConfig.Networks.Zones {
		zones = append(zones, map[string]interface{}{
			"name":     zone.Name,
			"worker":   zone.Workers,
			"public":   zone.Public,
			"internal": zone.Internal,
		})
	}
	return zones
}

func getSubnets(infrastructure *extensionsv1alpha1.Infrastructure, values map[string]string) ([]v1alpha1.Subnet, error) {
	var subnetsToReturn []v1alpha1.Subnet
	for key, value := range values {
		if strings.HasPrefix(key, aws.SubnetPublicPrefix) {
			zoneID, err := strconv.Atoi(strings.TrimPrefix(key, aws.SubnetPublicPrefix))
			if err != nil {
				return nil, err
			}
			subnetsToReturn = append(subnetsToReturn, v1alpha1.Subnet{
				Name: key,
				ID:   value,
				Zone: infrastructure.Spec.Zones[zoneID],
			})
		}
		if strings.HasPrefix(key, aws.SubnetNodesPrefix) {
			zoneID, err := strconv.Atoi(strings.TrimPrefix(key, aws.SubnetNodesPrefix))
			if err != nil {
				return nil, err
			}
			subnetsToReturn = append(subnetsToReturn, v1alpha1.Subnet{
				Name: key,
				ID:   value,
				Zone: infrastructure.Spec.Zones[zoneID],
			})
		}
	}
	return subnetsToReturn, nil
}
