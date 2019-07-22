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
	"fmt"

	"github.com/gardener/gardener-extensions/pkg/controller"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/util"
	"github.com/gardener/gardener-resource-manager/pkg/manager"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const mcmShootResourceName = "extension-worker-mcm-shoot"

func (a *genericActuator) deployMachineControllerManager(ctx context.Context, workerObj *extensionsv1alpha1.Worker, cluster *controller.Cluster, workerDelegate WorkerDelegate) error {
	mcmValues, err := workerDelegate.GetMachineControllerManagerChartValues(ctx)
	if err != nil {
		return err
	}

	// Generate MCM kubeconfig and inject its checksum into the MCM values.
	mcmKubeconfigSecret, err := createKubeconfigForMachineControllerManager(ctx, a.client, workerObj.Namespace, a.mcmName)
	if err != nil {
		return err
	}
	injectPodAnnotation(mcmValues, "checksum/secret-machine-controller-manager", util.ComputeChecksum(mcmKubeconfigSecret.Data))

	// If the shoot is hibernated then we want to scale down the machine-controller-manager. However, we want to first allow it to delete
	// all remaining worker nodes. Hence, we cannot set the replicas=0 here (otherwise it would be offline and not able to delete the nodes).
	if extensionscontroller.IsHibernated(cluster.Shoot) {
		deployment := &appsv1.Deployment{}
		if err := a.client.Get(ctx, kutil.Key(workerObj.Namespace, a.mcmName), deployment); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		if replicas := deployment.Spec.Replicas; replicas != nil {
			mcmValues["replicas"] = *replicas
		}
	}

	if err := a.mcmSeedChart.Apply(ctx, a.chartApplier, workerObj.Namespace,
		a.imageVector, a.gardenerClientset.Version(), cluster.Shoot.Spec.Kubernetes.Version, mcmValues); err != nil {
		return errors.Wrapf(err, "could not apply MCM chart in seed for worker '%s'", util.ObjectName(workerObj))
	}

	if !controller.IsHibernated(cluster.Shoot) {
		if err := a.applyMachineControllerManagerShootChart(ctx, workerDelegate, workerObj, cluster); err != nil {
			return err
		}
	}

	return nil
}

func (a *genericActuator) deleteMachineControllerManager(ctx context.Context, workerObj *extensionsv1alpha1.Worker) error {
	a.logger.Info("Deleting the machine-controller-manager", "worker", fmt.Sprintf("%s/%s", workerObj.Namespace, workerObj.Name))

	// Delete the managed resource
	a.logger.Info("Deleting managed resource containing mcm chart", "worker", util.ObjectName(workerObj), "name", mcmShootResourceName)
	if err := manager.NewManagedResource(a.client).
		WithNamespacedName(workerObj.Namespace, mcmShootResourceName).
		Delete(ctx); err != nil {
		return errors.Wrapf(err, "could not delete managed resource '%s/%s' containing mcm chart for worker '%s'", workerObj.Namespace, mcmShootResourceName, util.ObjectName(workerObj))
	}

	// Delete the secret referenced by the managed resource
	a.logger.Info("Deleting secret of managed resource containing mcm chart", "worker", util.ObjectName(workerObj), "name", mcmShootResourceName)
	if err := manager.NewSecret(a.client).
		WithNamespacedName(workerObj.Namespace, mcmShootResourceName).
		Delete(ctx); err != nil {
		return errors.Wrapf(err, "could not delete secret '%s/%s' of managed resource containing mcm chart for worker '%s'", workerObj.Namespace, mcmShootResourceName, util.ObjectName(workerObj))
	}

	if err := a.mcmSeedChart.Delete(ctx, a.client, workerObj.Namespace); err != nil {
		return errors.Wrapf(err, "cleaning up machine-controller-manager resources in seed failed")
	}

	return nil
}

func (a *genericActuator) applyMachineControllerManagerShootChart(ctx context.Context, workerDelegate WorkerDelegate, workerObj *extensionsv1alpha1.Worker, cluster *controller.Cluster) error {
	// Create shoot chart renderer
	chartRenderer, err := a.chartRendererFactory.NewChartRendererForShoot(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return errors.Wrapf(err, "could not create chart renderer for shoot '%s'", workerObj.Namespace)
	}

	// Get machine-controller-manager shoot chart values
	values, err := workerDelegate.GetMachineControllerManagerShootChartValues(ctx)
	if err != nil {
		return err
	}

	// Render machine-controller-manager shoot chart
	a.logger.Info("Rendering machine-controller-manager shoot chart", "worker", util.ObjectName(workerObj), "values", values)
	version := cluster.Shoot.Spec.Kubernetes.Version
	name, data, err := a.mcmShootChart.Render(chartRenderer, metav1.NamespaceSystem, a.imageVector, version, version, values)
	if err != nil {
		return errors.Wrapf(err, "could not render machine-controller-manager shoot chart for worker '%s'", util.ObjectName(workerObj))
	}

	// Create or update secret containing the rendered machine-controller-manager shoot chart
	a.logger.Info("Creating secret of managed resource containing mcm shoot chart", "worker", util.ObjectName(workerObj), "name", mcmShootResourceName)
	if err := manager.NewSecret(a.client).
		WithNamespacedName(workerObj.Namespace, mcmShootResourceName).
		WithKeyValues(map[string][]byte{name: data}).
		Reconcile(ctx); err != nil {
		return errors.Wrapf(err, "could not create or update secret '%s/%s' of managed resource containing mcm shoot chart for worker '%s'", workerObj.Namespace, mcmShootResourceName, util.ObjectName(workerObj))
	}

	// Create or update managed resource referencing the previously created secret
	a.logger.Info("Creating managed resource containing mcm shoot chart", "worker", util.ObjectName(workerObj), "name", mcmShootResourceName)
	if err := manager.NewManagedResource(a.client).
		WithNamespacedName(workerObj.Namespace, mcmShootResourceName).
		WithInjectedLabels(map[string]string{extensionscontroller.ShootNoCleanupLabel: "true"}).
		WithSecretRef(mcmShootResourceName).
		Reconcile(ctx); err != nil {
		return errors.Wrapf(err, "could not create or update managed resource '%s/%s' containing mcm shoot chart for worker '%s'", workerObj.Namespace, mcmShootResourceName, util.ObjectName(workerObj))
	}

	return nil
}

// createKubeconfigForMachineControllerManager generates a new certificate and kubeconfig for the machine-controller-manager. If
// such credentials already exist then they will be returned.
func createKubeconfigForMachineControllerManager(ctx context.Context, c client.Client, namespace, name string) (*corev1.Secret, error) {
	certConfig := secrets.CertificateSecretConfig{
		Name:       name,
		CommonName: fmt.Sprintf("system:%s", name),
	}

	return util.GetOrCreateShootKubeconfig(ctx, c, certConfig, namespace)
}

func injectPodAnnotation(values map[string]interface{}, key string, value interface{}) {
	podAnnotations, ok := values["podAnnotations"]
	if !ok {
		values["podAnnotations"] = map[string]interface{}{
			key: value,
		}
	} else {
		podAnnotations.(map[string]interface{})[key] = value
	}
}
