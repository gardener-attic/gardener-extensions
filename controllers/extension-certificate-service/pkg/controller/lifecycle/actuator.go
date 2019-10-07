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

package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config"
	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/controller/lifecycle/internal"
	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/imagevector"
	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/utils"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/chart"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/go-logr/logr"
	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Actuator acts upon Configuration.
type Actuator interface {
	// Reconcile the Configuration.
	Reconcile(ctx context.Context, config config.Configuration) error
}

// ActuatorName is the name of the Certificate Service Lifecycle actuator.
const ActuatorName = "certificate-service-lifecycle-actuator"

type actuator struct {
	client  client.Client
	config  *rest.Config
	applier kubernetes.ChartApplier

	logger logr.Logger
}

// NewActuator returns an actuator responsible for Configuration.
func NewActuator() Actuator {
	return &actuator{
		logger: log.Log.WithName(ActuatorName),
	}
}

// Create the Configuration.
func (a *actuator) Reconcile(ctx context.Context, config config.Configuration) error {
	return a.DeployCertManager(ctx, config)
}

// InjectConfig injects the rest config to this actuator.
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config
	applier, err := kubernetes.NewChartApplierForConfig(a.config)
	if err != nil {
		return fmt.Errorf("failed to create chart applier: %v", err)
	}
	a.applier = applier
	return nil
}

// InjectClient injects the client to this actuator.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

const certManagerName = "cert-manager"

var certManagerChartPath = filepath.Join(utils.ChartsPath, certManagerName)

func (a *actuator) DeployCertManager(ctx context.Context, config config.Configuration) error {
	var (
		applierOptions     = kubernetes.DefaultApplierOptions
		clusterIssuerMerge = func(new, old *unstructured.Unstructured) {
			// Apply status from old ClusterIssuer to retain the issuer's readiness state.
			new.Object["status"] = old.Object["status"]
		}
	)
	applierOptions.MergeFuncs[certmanagerv1alpha1.SchemeGroupVersion.WithKind("ClusterIssuer").GroupKind()] = clusterIssuerMerge

	namespace := corev1.Namespace{}
	if err := a.client.Get(ctx, kutil.Key(config.Spec.NamespaceRef), &namespace); err != nil {
		return fmt.Errorf("Failed fetching namespace %s for setting owner reference: %v", config.Spec.NamespaceRef, err)
	}

	configValues, err := internal.CreateCertServiceValues(config.Spec, namespace.Name, namespace.UID)
	if err != nil {
		return fmt.Errorf("failed to create values for cluster issuer: %v", err)
	}

	configValues, err = chart.InjectImages(configValues, imagevector.ImageVector(), []string{certManagerName})
	if err != nil {
		return fmt.Errorf("failed to find image version for %s: %v", certManagerName, err)
	}

	a.logger.Info("Component is being applied", "component", "cert-manager")
	if err := a.applier.ApplyChartInNamespaceWithOptions(
		ctx,
		certManagerChartPath,
		config.Spec.ResourceNamespace,
		certManagerName,
		configValues,
		nil,
		applierOptions,
	); err != nil {
		// TODO: (timuthy) This is an upgrade scenario from Gardener's in-tree deployment to the extension controller. Can be removed in the future.
		if apierrors.IsInvalid(err) {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      certManagerName,
					Namespace: config.Spec.ResourceNamespace,
				},
			}
			sa := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "seed-bootstrap-cert-manager",
					Namespace: config.Spec.ResourceNamespace,
				},
			}
			if err2 := a.client.Delete(ctx, deployment, kubernetes.ForceDeleteOptionFuncs...); err2 != nil {
				return fmt.Errorf("could not resolve deployment error '%s' by deleting cert-manager deployment '%s'", err, err2)
			}
			if err := a.client.Delete(ctx, sa); client.IgnoreNotFound(err) != nil {
				return err
			}
			return errors.New("old Cert-Manager deployment was found and deleted, Cert-Manager will be reinstalled in the next reconciliation loop")
		}
	}
	return nil
}
