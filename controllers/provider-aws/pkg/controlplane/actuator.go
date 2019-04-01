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

package controlplane

import (
	"context"

	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/controlplane/internal"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane"
	"github.com/gardener/gardener-extensions/pkg/util"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// NewActuator creates a new Actuator that acts upon and updates the status of ControlPlane resources.
func NewActuator(helper *internal.Helper) controlplane.Actuator {
	return &actuator{
		Helper: helper,
		logger: log.Log.WithName("controlplane-controller"),
	}
}

// actuator is an Actuator that acts upon and updates the status of ControlPlane resources.
type actuator struct {
	*internal.Helper
	decoder           runtime.Decoder
	clientset         kubernetes.Interface
	gardenerClientset gardenerkubernetes.Interface
	chartApplier      gardenerkubernetes.ChartApplier
	client            client.Client
	logger            logr.Logger
}

// InjectScheme injects the given scheme into the actuator.
func (a *actuator) InjectScheme(scheme *runtime.Scheme) error {
	a.decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
	return nil
}

// InjectConfig injects the given config into the actuator.
func (a *actuator) InjectConfig(config *rest.Config) error {
	// Create clientset
	var err error
	a.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "could not create Kubernetes client")
	}

	// Create Gardener clientset
	a.gardenerClientset, err = gardenerkubernetes.NewForConfig(config, client.Options{})
	if err != nil {
		return errors.Wrap(err, "could not create Gardener client")
	}

	// Create chart applier
	a.chartApplier, err = gardenerkubernetes.NewChartApplierForConfig(config)
	if err != nil {
		return errors.Wrap(err, "could not create chart applier")
	}

	return nil
}

// InjectClient injects the given client into the actuator.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

// Reconcile reconciles the given controlplane and cluster, creating or updating the additional Shoot
// control plane components as needed.
func (a *actuator) Reconcile(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) error {
	// Deploy secrets
	a.logger.Info("Deploying secrets", "controlplane", util.ObjectName(cp))
	deployedSecrets, err := a.Secrets.Deploy(a.clientset, a.gardenerClientset, cp.Namespace)
	if err != nil {
		return errors.Wrapf(err, "could not deploy secrets for controlplane '%s'", util.ObjectName(cp))
	}

	// Get config chart values
	values, err := a.ConfigChartValuesFunc(a.decoder, cp, cluster)
	if err != nil {
		return err
	}

	// Apply config chart
	a.logger.Info("Applying configuration chart", "controlplane", util.ObjectName(cp), "values", values)
	if err := a.ConfigChart.Apply(ctx, a.gardenerClientset, a.chartApplier, cp.Namespace, cluster.Shoot, nil, nil, values); err != nil {
		return errors.Wrapf(err, "could not apply configuration chart for controlplane '%s'", util.ObjectName(cp))
	}

	// Compute all needed checksums
	checksums, err := controlplane.ComputeChecksums(ctx, a.client, deployedSecrets, nil,
		[]string{common.CloudProviderSecretName}, []string{common.CloudProviderConfigName}, cp.Namespace)
	if err != nil {
		return errors.Wrapf(err, "could not compute checksums for controlplane '%s'", util.ObjectName(cp))
	}

	// Get control plane chart values
	values, err = a.ControlPlaneChartValuesFunc(a.decoder, cp, cluster, checksums)
	if err != nil {
		return err
	}

	// Apply control plane chart
	a.logger.Info("Applying control plane chart", "controlplane", util.ObjectName(cp), "values", values)
	if err := a.ControlPlaneChart.Apply(ctx, a.gardenerClientset, a.chartApplier, cp.Namespace, cluster.Shoot, a.ImageVectorFunc(), checksums, values); err != nil {
		return errors.Wrapf(err, "could not apply control plane chart for controlplane '%s'", util.ObjectName(cp))
	}

	return nil
}

// Delete reconciles the given controlplane and cluster, deleting the additional Shoot control plane components as needed.
func (a *actuator) Delete(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) error {
	// Delete control plane objects
	a.logger.Info("Deleting control plane objects", "controlplane", util.ObjectName(cp))
	if err := a.ControlPlaneChart.Delete(ctx, a.client, cp.Namespace); err != nil {
		return errors.Wrapf(err, "could not delete control plane objects for controlplane '%s'", util.ObjectName(cp))
	}

	// Delete config objects
	a.logger.Info("Deleting configuration objects", "controlplane", util.ObjectName(cp))
	if err := a.ConfigChart.Delete(ctx, a.client, cp.Namespace); err != nil {
		return errors.Wrapf(err, "could not delete configuration objects for controlplane '%s'", util.ObjectName(cp))
	}

	// Delete secrets
	a.logger.Info("Deleting secrets", "controlplane", util.ObjectName(cp))
	if err := a.Secrets.Delete(a.clientset, cp.Namespace); err != nil {
		return errors.Wrapf(err, "could not delete secrets for controlplane '%s'", util.ObjectName(cp))
	}

	return nil
}
