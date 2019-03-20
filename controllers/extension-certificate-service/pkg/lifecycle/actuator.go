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
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/lifecycle/internal"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/imagevector"

	configv1alpha1 "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/utils"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/go-logr/logr"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/rest"
)

// Actuator acts upon CertificateServiceConfiguration.
type Actuator interface {
	// CreateOrUpdate the CertificateServiceConfiguration.
	CreateOrUpdate(ctx context.Context, config *configv1alpha1.CertificateServiceConfiguration) error
	// Delete the CertificateServiceConfiguration.
	Delete(ctx context.Context, config *configv1alpha1.CertificateServiceConfiguration) error
}

// ActuatorName is the name of the Certificate Service Lifecycle actuator.
const ActuatorName = "certificate-service-lifecycle-actuator"

type actuator struct {
	client            client.Client
	config            *rest.Config
	resourceNamespace string
	scheme            *runtime.Scheme

	logger logr.Logger
}

// NewActuator returns an actuator responsible for CertificateServiceConfiguration.
func NewActuator(namespace string) Actuator {
	return &actuator{
		logger:            log.Log.WithName(ActuatorName),
		resourceNamespace: namespace,
	}
}

// Create the CertificateServiceConfiguration.
func (a *actuator) CreateOrUpdate(ctx context.Context, config *configv1alpha1.CertificateServiceConfiguration) error {
	return a.DeployCertManager(ctx, config)
}

// Delete the CertificateServiceConfiguration.
func (a *actuator) Delete(ctx context.Context, config *configv1alpha1.CertificateServiceConfiguration) error {
	return a.DeleteCertManager(ctx, config)
}

// InjectConfig injects the rest config to this actuator.
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config
	return nil
}

// InjectConfig injects the scheme to this actuator.
func (a *actuator) InjectScheme(scheme *runtime.Scheme) error {
	a.scheme = scheme
	return nil
}

// InjectClient injects the client to this actuator.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

const certManagerName = "cert-manager"

var certManagerChartPath = filepath.Join(utils.ChartsPath, certManagerName)

func (a *actuator) DeployCertManager(ctx context.Context, config *configv1alpha1.CertificateServiceConfiguration) error {
	applier, err := kubernetes.NewChartApplierForConfig(a.config)
	if err != nil {
		return fmt.Errorf("failed to create chart applier: %v", err)
	}

	var (
		applierOptions     = kubernetes.DefaultApplierOptions
		clusterIssuerMerge = func(new, old *unstructured.Unstructured) {
			// Apply status from old ClusterIssuer to retain the issuer's readiness state.
			new.Object["status"] = old.Object["status"]
		}
	)
	applierOptions.MergeFuncs["ClusterIssuer"] = clusterIssuerMerge

	configValues, err := internal.CreateCertServiceValues(config.Spec)
	if err != nil {
		return fmt.Errorf("failed to create values for cluster issuer: %v", err)
	}

	configValues, err = imagevector.ImageVector.InjectImages(configValues, "", "", certManagerName)
	if err != nil {
		return fmt.Errorf("failed to find image version for %s: %v", certManagerName, err)
	}

	a.logger.Info("Component is being applied", "component", "cert-manager")
	return applier.ApplyChartInNamespaceWithOptions(
		ctx,
		certManagerChartPath,
		a.resourceNamespace,
		certManagerName,
		configValues,
		nil,
		applierOptions,
	)
}

func (a *actuator) DeleteCertManager(ctx context.Context, config *configv1alpha1.CertificateServiceConfiguration) error {
	chartRenderer, err := chartrenderer.NewForConfig(a.config)
	if err != nil {
		return err
	}
	configValues, err := internal.CreateCertServiceValues(config.Spec)
	if err != nil {
		return fmt.Errorf("failed to create values for cluster issuer: %v", err)
	}
	release, err := chartRenderer.Render(
		certManagerChartPath,
		certManagerName,
		a.resourceNamespace,
		configValues)
	if err != nil {
		return fmt.Errorf("failed to render chart: %v", err)
	}

	var (
		manifest   = release.Manifest()
		decoder    = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 1024)
		decodedObj map[string]interface{}
	)

	a.logger.Info("Component is being deleted", "component", "cert-manager")
	for err = decoder.Decode(&decodedObj); err == nil; err = decoder.Decode(&decodedObj) {
		if decodedObj == nil {
			continue
		}

		obj := unstructured.Unstructured{Object: decodedObj}
		decodedObj = nil

		// Do not delete CustomResourceDefinitions to not interfere with other resources of that kind.
		if obj.GroupVersionKind() != apiextensionsv1beta1.SchemeGroupVersion.WithKind("CustomResourceDefinition") {
			if err := a.client.Delete(ctx, &obj); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}

	}

	return nil
}
