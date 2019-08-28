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

package terraformer

import (
	"time"

	gardenerterraformer "github.com/gardener/gardener/pkg/operation/terraformer"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type terraformer struct {
	tf *gardenerterraformer.Terraformer
}

// SetVariablesEnvironment implements Terraformer.
func (t *terraformer) SetVariablesEnvironment(tfVarsEnvironment map[string]string) Interface {
	return &terraformer{t.tf.SetVariablesEnvironment(tfVarsEnvironment)}
}

// SetJobBackoffLimit implements Terraformer.
func (t *terraformer) SetJobBackoffLimit(val int32) Interface {
	return &terraformer{t.tf.SetJobBackoffLimit(val)}
}

// SetActiveDeadlineSeconds implements Terraformer.
func (t *terraformer) SetActiveDeadlineSeconds(val int64) Interface {
	return &terraformer{t.tf.SetActiveDeadlineSeconds(val)}
}

// SetDeadlineCleaning implements Terraformer.
func (t *terraformer) SetDeadlineCleaning(val time.Duration) Interface {
	return &terraformer{t.tf.SetDeadlineCleaning(val)}
}

// SetDeadlinePod implements Terraformer.
func (t *terraformer) SetDeadlinePod(val time.Duration) Interface {
	return &terraformer{t.tf.SetDeadlinePod(val)}
}

// SetDeadlineJob implements Terraformer.
func (t *terraformer) SetDeadlineJob(val time.Duration) Interface {
	return &terraformer{t.tf.SetDeadlineJob(val)}
}

// InitializeWith implements Terraformer.
func (t *terraformer) InitializeWith(initializer Initializer) Interface {
	return &terraformer{t.tf.InitializeWith(initializer.Initialize)}
}

// Apply implements Terraformer.
func (t *terraformer) Apply() error {
	return t.tf.Apply()
}

// Destroy implements Terraformer.
func (t *terraformer) Destroy() error {
	return t.tf.Destroy()
}

// GetStateOutputVariables implements Terraformer.
func (t *terraformer) GetStateOutputVariables(variables ...string) (map[string]string, error) {
	return t.tf.GetStateOutputVariables(variables...)
}

// ConfigExists implements Terraformer.
func (t *terraformer) ConfigExists() (bool, error) {
	return t.tf.ConfigExists()
}

type initializerFunc func(config *gardenerterraformer.InitializerConfig) error

// Initialize implements Initializer.
func (f initializerFunc) Initialize(config *gardenerterraformer.InitializerConfig) error {
	return f(config)
}

type factory struct{}

// NewForConfig implements Factory.
func (factory) NewForConfig(logger logrus.FieldLogger, config *rest.Config, purpose, namespace, name, image string) (Interface, error) {
	tf, err := gardenerterraformer.NewForConfig(logger, config, purpose, namespace, name, image)
	if err != nil {
		return nil, err
	}

	return &terraformer{tf}, nil
}

// New implements Factory.
func (factory) New(logger logrus.FieldLogger, client client.Client, coreV1Client v1.CoreV1Interface, purpose, namespace, name, image string) Interface {
	return &terraformer{gardenerterraformer.New(logger, client, coreV1Client, purpose, namespace, name, image)}
}

// DefaultInitializer implements Factory.
func (factory) DefaultInitializer(c client.Client, main, variables string, tfVars []byte) Initializer {
	return initializerFunc(gardenerterraformer.DefaultInitializer(c, main, variables, tfVars))
}

// DefaultFactory returns the default factory.
func DefaultFactory() Factory {
	return factory{}
}
