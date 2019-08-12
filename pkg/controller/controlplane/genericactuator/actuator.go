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

package genericactuator

import (
	"context"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/controlplane"
	"github.com/gardener/gardener-extensions/pkg/util"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// ValuesProvider provides values for the 2 charts applied by this actuator.
type ValuesProvider interface {
	// GetConfigChartValues returns the values for the config chart applied by this actuator.
	GetConfigChartValues(context.Context, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster) (map[string]interface{}, error)
	// GetControlPlaneChartValues returns the values for the control plane chart applied by this actuator.
	GetControlPlaneChartValues(context.Context, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster, map[string]string, bool) (map[string]interface{}, error)
	// GetControlPlaneShootChartValues returns the values for the control plane shoot chart applied by this actuator.
	GetControlPlaneShootChartValues(context.Context, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster) (map[string]interface{}, error)
	// GetStorageClassesChartValues returns the values for the storage classes chart applied by this actuator.
	GetStorageClassesChartValues(context.Context, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster) (map[string]interface{}, error)
}

// NewActuator creates a new Actuator that acts upon and updates the status of ControlPlane resources.
// It creates / deletes the given secrets and applies / deletes the given charts, using the given image vector and
// the values provided by the given values provider.
func NewActuator(
	secrets util.Secrets,
	configChart, controlPlaneChart, controlPlaneShootChart, storageClassesChart util.Chart,
	vp ValuesProvider,
	chartRendererFactory extensionscontroller.ChartRendererFactory,
	imageVector imagevector.ImageVector,
	configName string,
	logger logr.Logger,
) controlplane.Actuator {
	return &actuator{
		secrets:                secrets,
		configChart:            configChart,
		controlPlaneChart:      controlPlaneChart,
		controlPlaneShootChart: controlPlaneShootChart,
		storageClassesChart:    storageClassesChart,
		vp:                     vp,
		chartRendererFactory:   chartRendererFactory,
		imageVector:            imageVector,
		configName:             configName,
		logger:                 logger.WithName("controlplane-actuator"),
	}
}

// actuator is an Actuator that acts upon and updates the status of ControlPlane resources.
type actuator struct {
	secrets                util.Secrets
	configChart            util.Chart
	controlPlaneChart      util.Chart
	controlPlaneShootChart util.Chart
	storageClassesChart    util.Chart
	vp                     ValuesProvider
	chartRendererFactory   extensionscontroller.ChartRendererFactory
	imageVector            imagevector.ImageVector
	configName             string

	clientset         kubernetes.Interface
	gardenerClientset gardenerkubernetes.Interface
	chartApplier      gardenerkubernetes.ChartApplier
	client            client.Client
	logger            logr.Logger
}

// InjectFunc enables injecting Kubernetes dependencies into actuator's dependencies.
func (a *actuator) InjectFunc(f inject.Func) error {
	return f(a.vp)
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

// InjectClient injects the given client into the valuesProvider.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

const (
	controlPlaneShootChartResourceName = "extension-controlplane-shoot"
	storageClassesChartResourceName    = "extension-controlplane-storageclasses"
)

// Reconcile reconciles the given controlplane and cluster, creating or updating the additional Shoot
// control plane components as needed.
func (a *actuator) Reconcile(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) (bool, error) {
	// Deploy secrets
	a.logger.Info("Deploying secrets", "controlplane", util.ObjectName(cp))
	deployedSecrets, err := a.secrets.Deploy(a.clientset, a.gardenerClientset, cp.Namespace)
	if err != nil {
		return false, errors.Wrapf(err, "could not deploy secrets for controlplane '%s'", util.ObjectName(cp))
	}

	// Get config chart values
	if a.configChart != nil {
		values, err := a.vp.GetConfigChartValues(ctx, cp, cluster)
		if err != nil {
			return false, err
		}

		// Apply config chart
		a.logger.Info("Applying configuration chart", "controlplane", util.ObjectName(cp))
		if err := a.configChart.Apply(ctx, a.chartApplier, cp.Namespace, nil, "", "", values); err != nil {
			return false, errors.Wrapf(err, "could not apply configuration chart for controlplane '%s'", util.ObjectName(cp))
		}
	}

	// Compute all needed checksums
	checksums, err := a.computeChecksums(ctx, deployedSecrets, cp.Namespace)
	if err != nil {
		return false, err
	}

	var (
		requeue    = false
		scaledDown = false
	)

	if extensionscontroller.IsHibernated(cluster.Shoot) {
		dep := &appsv1.Deployment{}
		if err := a.client.Get(ctx, client.ObjectKey{Namespace: cp.Namespace, Name: gardencorev1alpha1.DeploymentNameKubeAPIServer}, dep); err != nil {
			return false, errors.Wrapf(err, "could not get deployment '%s/%s'", cp.Namespace, gardencorev1alpha1.DeploymentNameKubeAPIServer)
		}

		// If the cluster is hibernated, check if kube-apiserver has been already scaled down. If it is not yet scaled down
		// then we requeue the `ControlPlane` CRD in order to give the provider-specific control plane components time to
		// properly prepare the cluster for hibernation (whatever needs to be done). If the kube-apiserver is already scaled down
		// then we allow continuing the reconciliation.
		if cluster.Shoot.DeletionTimestamp == nil {
			if dep.Spec.Replicas != nil && *dep.Spec.Replicas > 0 {
				requeue = true
			} else {
				scaledDown = true
			}
			// Similarly, if a hibernated shoot is deleted then we might need to wake up all the provider-specific components. We
			// wait until the kube-apiserver is woken up again before we wake up the provider-specific components.
		} else {
			if dep.Spec.Replicas == nil || *dep.Spec.Replicas == 0 {
				return true, nil
			}
		}
	}

	// Get control plane chart values
	values, err := a.vp.GetControlPlaneChartValues(ctx, cp, cluster, checksums, scaledDown)
	if err != nil {
		return false, err
	}

	// Apply control plane chart
	a.logger.Info("Applying control plane chart", "controlplane", util.ObjectName(cp))
	if err := a.controlPlaneChart.Apply(ctx, a.chartApplier, cp.Namespace, a.imageVector, a.gardenerClientset.Version(), cluster.Shoot.Spec.Kubernetes.Version, values); err != nil {
		return false, errors.Wrapf(err, "could not apply control plane chart for controlplane '%s'", util.ObjectName(cp))
	}

	// Create shoot chart renderer
	chartRenderer, err := a.chartRendererFactory.NewChartRendererForShoot(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return false, errors.Wrapf(err, "could not create chart renderer for shoot '%s'", cp.Namespace)
	}

	// Get control plane shoot chart values
	values, err = a.vp.GetControlPlaneShootChartValues(ctx, cp, cluster)
	if err != nil {
		return false, err
	}

	if err := extensionscontroller.RenderChartAndCreateManagedResource(ctx, cp.Namespace, controlPlaneShootChartResourceName, a.client, chartRenderer, a.controlPlaneShootChart, values, a.imageVector, metav1.NamespaceSystem, cluster.Shoot.Spec.Kubernetes.Version, true); err != nil {
		return false, errors.Wrapf(err, "could not apply control plane shoot chart for controlplane '%s'", util.ObjectName(cp))
	}

	// Get storage classes
	values, err = a.vp.GetStorageClassesChartValues(ctx, cp, cluster)
	if err != nil {
		return false, err
	}

	if err := extensionscontroller.RenderChartAndCreateManagedResource(ctx, cp.Namespace, storageClassesChartResourceName, a.client, chartRenderer, a.storageClassesChart, values, a.imageVector, metav1.NamespaceSystem, cluster.Shoot.Spec.Kubernetes.Version, true); err != nil {
		return false, errors.Wrapf(err, "could not apply storage classes chart for controlplane '%s'", util.ObjectName(cp))
	}

	return requeue, nil
}

// Delete reconciles the given controlplane and cluster, deleting the additional Shoot
// control plane components as needed.
func (a *actuator) Delete(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) error {
	// Delete the managed resources
	if err := extensionscontroller.DeleteManagedResource(ctx, a.client, cp.Namespace, storageClassesChartResourceName); err != nil {
		return errors.Wrapf(err, "could not delete managed resource containing storage classes chart for controlplane '%s'", util.ObjectName(cp))
	}

	if err := extensionscontroller.DeleteManagedResource(ctx, a.client, cp.Namespace, controlPlaneShootChartResourceName); err != nil {
		return errors.Wrapf(err, "could not delete managed resource containing shoot chart for controlplane '%s'", util.ObjectName(cp))
	}

	// Delete control plane objects
	a.logger.Info("Deleting control plane objects", "controlplane", util.ObjectName(cp))
	if err := a.controlPlaneChart.Delete(ctx, a.client, cp.Namespace); err != nil {
		return errors.Wrapf(err, "could not delete control plane objects for controlplane '%s'", util.ObjectName(cp))
	}

	if a.configChart != nil {
		// Delete config objects
		a.logger.Info("Deleting configuration objects", "controlplane", util.ObjectName(cp))
		if err := a.configChart.Delete(ctx, a.client, cp.Namespace); err != nil {
			return errors.Wrapf(err, "could not delete configuration objects for controlplane '%s'", util.ObjectName(cp))
		}
	}

	// Delete secrets
	a.logger.Info("Deleting secrets", "controlplane", util.ObjectName(cp))
	if err := a.secrets.Delete(a.clientset, cp.Namespace); err != nil {
		return errors.Wrapf(err, "could not delete secrets for controlplane '%s'", util.ObjectName(cp))
	}

	return nil
}

// computeChecksums computes and returns all needed checksums. This includes the checksums for the given deployed secrets,
// as well as the cloud provider secret and configmap that are fetched from the cluster.
func (a *actuator) computeChecksums(
	ctx context.Context,
	deployedSecrets map[string]*corev1.Secret,
	namespace string,
) (map[string]string, error) {
	// Get cloud provider secret and config from cluster
	cpSecret := &corev1.Secret{}
	if err := a.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: gardencorev1alpha1.SecretNameCloudProvider}, cpSecret); err != nil {
		return nil, errors.Wrapf(err, "could not get secret '%s/%s'", namespace, gardencorev1alpha1.SecretNameCloudProvider)
	}

	csSecrets := controlplane.MergeSecretMaps(deployedSecrets, map[string]*corev1.Secret{
		gardencorev1alpha1.SecretNameCloudProvider: cpSecret,
	})

	var csConfigMaps map[string]*corev1.ConfigMap
	if len(a.configName) != 0 {
		cpConfigMap := &corev1.ConfigMap{}
		if err := a.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: a.configName}, cpConfigMap); err != nil {
			return nil, errors.Wrapf(err, "could not get configmap '%s/%s'", namespace, a.configName)
		}

		csConfigMaps = map[string]*corev1.ConfigMap{
			a.configName: cpConfigMap,
		}
	}

	return controlplane.ComputeChecksums(csSecrets, csConfigMaps), nil
}
