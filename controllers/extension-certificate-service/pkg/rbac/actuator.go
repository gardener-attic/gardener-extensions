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

package rbac

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/utils"

	"github.com/gardener/gardener/pkg/client/kubernetes"

	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// Actuator acts upon Secrets.
type Actuator interface {
	// Create the Secret.
	Create(ctx context.Context) error
	// Delete the Secret.
	Delete(ctx context.Context) error
}

// ActuatorName is the name of the Certificate Service RBAC actuator.
const ActuatorName = "certificate-service-rbac-actuator"

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator() Actuator {
	return &actuator{
		logger: log.Log.WithName(ActuatorName),
	}
}

type actuator struct {
	client client.Client
	config *rest.Config

	logger logr.Logger
}

var chartPath = filepath.Join(utils.ChartsPath, "cert-broker-rbac")

func (a *actuator) Create(ctx context.Context) error {
	applier, err := kubernetes.NewChartApplierForConfig(a.config)
	if err != nil {
		return fmt.Errorf("failed to create chart applier: %v", err)
	}

	a.logger.Info("Manifests are being applied")
	if err := applier.ApplyChart(ctx, chartPath, "", "", nil, nil); err != nil {
		return fmt.Errorf("error applying manifests: %v", err)
	}

	return nil
}

func (a *actuator) Delete(ctx context.Context) error {
	renderer, err := chartrenderer.NewForConfig(a.config)
	if err != nil {
		return fmt.Errorf("failed to create chart renderer for rbac removal: %v", err)
	}

	release, err := renderer.Render(chartPath, "", "", nil)
	if err != nil {
		return fmt.Errorf("failed to render charts for rbac removal: %v", err)
	}

	var (
		manifest   = release.Manifest()
		decoder    = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 1024)
		decodedObj map[string]interface{}
	)

	a.logger.Info("Manifests are being deleted")
	for err = decoder.Decode(&decodedObj); err == nil; err = decoder.Decode(&decodedObj) {
		if decodedObj == nil {
			continue
		}

		obj := unstructured.Unstructured{Object: decodedObj}
		decodedObj = nil

		if err := a.client.Delete(ctx, &obj); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// InjectClient injects the controller runtime client into the reconciler.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

// InjectConfig injects the rest config to this actuator.
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config
	return nil
}
