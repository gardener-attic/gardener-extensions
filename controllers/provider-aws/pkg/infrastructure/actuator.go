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
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/gardener/gardener/pkg/client/aws"

	"github.com/gardener/gardener/pkg/chartrenderer"

	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"

	corev1 "k8s.io/api/core/v1"

	awstypes "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws/v1alpha1"

	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/terraformer"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type actuator struct {
	config     *rest.Config
	client     client.Client
	kubernetes kubernetes.Interface
	scheme     *runtime.Scheme
	logger     logr.Logger
	decoder    runtime.Decoder
	encoder    runtime.Encoder
}

// NewActuator creates a new Actuator that updates the status of the handled WorkerPoolConfigs.
func NewActuator() infrastructure.Actuator {
	return &actuator{logger: log.Log.WithName("infrastructure-aws-actuator")}
}

func (c *actuator) InjectScheme(scheme *runtime.Scheme) error {
	c.scheme = scheme
	c.decoder = serializer.NewCodecFactory(c.scheme).UniversalDecoder()
	c.encoder = serializer.NewCodecFactory(c.scheme).EncoderForVersion(c.encoder, extensionsv1alpha1.SchemeGroupVersion)
	return nil
}

func (c *actuator) InjectConfig(config *rest.Config) error {
	c.config = config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	c.kubernetes = clientset
	return nil
}

func (c *actuator) InjectClient(client client.Client) error {
	c.client = client
	return nil
}

func (c *actuator) Exists(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure) (bool, error) {
	return infrastructure.Status.LastOperation != nil, nil
}

func (c *actuator) Create(ctx context.Context, config *extensionsv1alpha1.Infrastructure) error {
	return c.reconcile(ctx, config)
}

func (c *actuator) Update(ctx context.Context, config *extensionsv1alpha1.Infrastructure) error {
	return c.reconcile(ctx, config)
}

func (c *actuator) Delete(ctx context.Context, config *extensionsv1alpha1.Infrastructure) error {
	return c.delete(ctx, config)
}

func (c *actuator) reconcile(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure) error {
	infrastructureConfig := &v1alpha1.InfrastructureConfig{}
	_, _, err := c.decoder.Decode(infrastructure.Spec.ProviderConfig.Raw, nil, infrastructureConfig)
	if err != nil {
		c.injectStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "could not decode providerConfig", 10, err)
		return err
	}

	var (
		namespace = infrastructure.Spec.SecretRef.Namespace
		name      = infrastructure.Spec.SecretRef.Name
	)

	providerSecret := &corev1.Secret{}
	if err := c.client.Get(ctx, kutil.Key(namespace, name), providerSecret); err != nil {
		return err
	}
	var (
		createVPC         = true
		vpcID             = "${aws_vpc.vpc.id}"
		internetGatewayID = "${aws_internet_gateway.igw.id}"
		dhcpDomainName    = "ec2.internal"
		vpcCIDR           = infrastructureConfig.Networks.VPC.CIDR
		region            = infrastructure.Spec.Region
	)

	awsClient := aws.NewClient(string(providerSecret.Data[awstypes.AccessKeyID]), string(providerSecret.Data[awstypes.SecretAccessKey]), region)

	// check if we should use an existing VPC or create a new one
	if len(infrastructureConfig.Networks.VPC.ID) > 0 {
		createVPC = false
		vpcID = infrastructureConfig.Networks.VPC.ID
		igwID, err := awsClient.GetInternetGateway(vpcID)
		if err != nil {
			return err
		}
		internetGatewayID = igwID
	} else if len(infrastructureConfig.Networks.VPC.CIDR) > 0 {
		vpcCIDR = infrastructureConfig.Networks.VPC.CIDR
	}

	values := map[string]interface{}{
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
		"clusterName": infrastructure.Spec.SecretRef.Namespace,
		"zones":       getZones(infrastructure, infrastructureConfig),
	}

	chartRenderer, err := chartrenderer.New(c.kubernetes)
	if err != nil {
		c.injectStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "could not create chart renderer", 30, err)
		return err
	}

	release, err := chartRenderer.Render(filepath.Join(awstypes.TerraformersChartsPath, "aws-infra"), "aws-infra", namespace, values)
	if err != nil {
		c.injectStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "could not render chart", 50, err)
		return err
	}

	tf, err := c.getTerraformer(awstypes.TerrformerPurposeInfra, infrastructure.Namespace, infrastructure.Name)
	if err != nil {
		c.injectStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "could not create terraformer object", 70, err)
		return err
	}

	if err = tf.SetVariablesEnvironment(terraformer.GenerateVariablesEnvironment(providerSecret, map[string]string{
		"ACCESS_KEY_ID":     awstypes.AccessKeyID,
		"SECRET_ACCESS_KEY": awstypes.SecretAccessKey,
	})).InitializeWith(terraformer.DefaultInitializer(c.client,
		release.FileContent("main.tf"),
		release.FileContent("variables.tf"),
		[]byte(release.FileContent("terraform.tfvars")))).Apply(); err != nil {
		c.injectStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "could not apply terraform config", 80, err)
		return err
	}

	if err = c.injectProviderStateIntoStatus(context.TODO(), tf, infrastructure); err != nil {
		c.injectStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "failed to update status with the provider state", 90, err)
		return err
	}

	return c.injectStatusSuccess(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeReconcile, "successfully reconciled infrastructure")
}

func (c *actuator) delete(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure) error {
	c.logger.Info("Destroying Shoot infrastructure")

	tf, err := c.getTerraformer(awstypes.TerrformerPurposeInfra, infrastructure.Namespace, infrastructure.Name)
	if err != nil {
		c.injectStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeDelete, "could not create terraformer object", 30, err)
		return err
	}

	configExists, err := tf.ConfigExists()
	if err != nil {
		c.injectStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeDelete, "terraform configuration was not found on seed", 60, err)
		return err
	}

	if configExists {
		var (
			namespace = infrastructure.Spec.SecretRef.Namespace
			name      = infrastructure.Spec.SecretRef.Name
			region    = infrastructure.Spec.Region
		)

		providerSecret := &corev1.Secret{}
		if err := c.client.Get(ctx, kutil.Key(namespace, name), providerSecret); err != nil {
			return err
		}

		awsClient := aws.NewClient(string(providerSecret.Data[awstypes.AccessKeyID]), string(providerSecret.Data[awstypes.SecretAccessKey]), region)

		c.logger.Info("Destroying Kubernetes load balancers and security groups")
		err := c.destroyKubernetesLoadBalancersAndSecurityGroups(namespace, tf, awsClient)
		if err != nil {
			c.injectStatusError(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeDelete, "failed to destroy kubernetes load balancers and security groups", 80, err)
		}
	}

	return c.injectStatusSuccess(ctx, infrastructure, extensionsv1alpha1.LastOperationTypeDelete, "successfully deleted infrastructure")
}
