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
	gardenerterraformer "github.com/gardener/gardener/pkg/operation/terraformer"
	"github.com/sirupsen/logrus"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Interface is the terraformer interface.
type Interface interface {
	SetVariablesEnvironment(tfVarsEnvironment map[string]string) Interface
	InitializeWith(initializer Initializer) Interface
	Apply() error
	Destroy() error
	GetStateOutputVariables(variables ...string) (map[string]string, error)
	ConfigExists() (bool, error)
}

// Factory is a factory that can produce Interface and Initializer.
type Factory interface {
	NewForConfig(logger logrus.FieldLogger, config *rest.Config, purpose, namespace, name, image string) (Interface, error)
	New(logger logrus.FieldLogger, client client.Client, coreV1Client corev1client.CoreV1Interface, purpose, namespace, name, image string) Interface
	DefaultInitializer(c client.Client, main, variables string, tfVars []byte) Initializer
}

// Initializer can initialize an Interface.
type Initializer interface {
	Initialize(config *gardenerterraformer.InitializerConfig) error
}
