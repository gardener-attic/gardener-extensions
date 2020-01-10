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
	"github.com/gardener/gardener-extensions/pkg/controller/worker"
	"github.com/gardener/gardener-extensions/pkg/util"

	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

type genericActuator struct {
	logger logr.Logger

	delegateFactory DelegateFactory
	mcmName         string
	mcmSeedChart    util.Chart
	mcmShootChart   util.Chart
	imageVector     imagevector.ImageVector

	client               client.Client
	clientset            kubernetes.Interface
	decoder              runtime.Decoder
	gardenerClientset    gardenerkubernetes.Interface
	chartApplier         gardenerkubernetes.ChartApplier
	chartRendererFactory extensionscontroller.ChartRendererFactory
}

// NewActuator creates a new Actuator that reconciles
// Worker resources of Gardener's `extensions.gardener.cloud` API group.
// It provides a default implementation that allows easier integration of providers.
func NewActuator(logger logr.Logger, delegateFactory DelegateFactory, mcmName string, mcmSeedChart, mcmShootChart util.Chart, imageVector imagevector.ImageVector, chartRendererFactory extensionscontroller.ChartRendererFactory) worker.Actuator {
	return &genericActuator{
		logger: logger.WithName("worker-actuator"),

		delegateFactory:      delegateFactory,
		mcmName:              mcmName,
		mcmSeedChart:         mcmSeedChart,
		mcmShootChart:        mcmShootChart,
		imageVector:          imageVector,
		chartRendererFactory: chartRendererFactory,
	}
}

func (a *genericActuator) InjectFunc(f inject.Func) error {
	return f(a.delegateFactory)
}

func (a *genericActuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

func (a *genericActuator) InjectScheme(scheme *runtime.Scheme) error {
	a.decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
	return nil
}

func (a *genericActuator) InjectConfig(config *rest.Config) error {
	var err error

	a.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "could not create Kubernetes client")
	}

	a.gardenerClientset, err = gardenerkubernetes.NewWithConfig(gardenerkubernetes.WithRESTConfig(config))
	if err != nil {
		return errors.Wrap(err, "could not create Gardener client")
	}

	a.chartApplier, err = gardenerkubernetes.NewChartApplierForConfig(config)
	if err != nil {
		return errors.Wrap(err, "could not create chart applier")
	}

	return nil
}

func (a *genericActuator) cleanupMachineDeployments(ctx context.Context, existingMachineDeployments *machinev1alpha1.MachineDeploymentList, wantedMachineDeployments worker.MachineDeployments) error {
	for _, existingMachineDeployment := range existingMachineDeployments.Items {
		if !wantedMachineDeployments.HasDeployment(existingMachineDeployment.Name) {
			if err := a.client.Delete(ctx, &existingMachineDeployment); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *genericActuator) listMachineClassNames(ctx context.Context, namespace string, machineClassList runtime.Object) (sets.String, error) {
	if err := a.client.List(ctx, machineClassList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	classNames := sets.NewString()

	if err := meta.EachListItem(machineClassList, func(machineClass runtime.Object) error {
		accessor, err := meta.Accessor(machineClass)
		if err != nil {
			return err
		}

		classNames.Insert(accessor.GetName())
		return nil
	}); err != nil {
		return nil, err
	}

	return classNames, nil
}

func (a *genericActuator) cleanupMachineClasses(ctx context.Context, namespace string, machineClassList runtime.Object, wantedMachineDeployments worker.MachineDeployments) error {
	if err := a.client.List(ctx, machineClassList, client.InNamespace(namespace)); err != nil {
		return err
	}

	return meta.EachListItem(machineClassList, func(machineClass runtime.Object) error {
		accessor, err := meta.Accessor(machineClass)
		if err != nil {
			return err
		}

		if !wantedMachineDeployments.HasClass(accessor.GetName()) {
			if err := a.client.Delete(ctx, machineClass); err != nil {
				return err
			}
		}

		return nil
	})
}

func (a *genericActuator) listMachineClassSecrets(ctx context.Context, namespace string) (*corev1.SecretList, error) {
	var (
		secretList = &corev1.SecretList{}
		labels     = map[string]string{
			v1beta1constants.GardenPurpose: v1beta1constants.GardenPurposeMachineClass,
		}
	)

	if err := a.client.List(ctx, secretList, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
		return nil, err
	}
	return secretList, nil
}

// cleanupMachineClassSecrets deletes all unused machine class secrets (i.e., those which are not part
// of the provided list <usedSecrets>.
func (a *genericActuator) cleanupMachineClassSecrets(ctx context.Context, namespace string, wantedMachineDeployments worker.MachineDeployments) error {
	secretList, err := a.listMachineClassSecrets(ctx, namespace)
	if err != nil {
		return err
	}

	// Cleanup all secrets which were used for machine classes that do not exist anymore.
	for _, secret := range secretList.Items {
		if !wantedMachineDeployments.HasSecret(secret.Name) {
			if err := a.client.Delete(ctx, &secret); err != nil {
				return err
			}
		}
	}

	return nil
}
