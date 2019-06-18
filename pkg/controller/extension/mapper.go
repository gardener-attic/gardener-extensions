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

package extension

import (
	"github.com/gardener/gardener-extensions/pkg/controller"
	extensions1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ClusterToExtensionMapper returns a mapper that returns requests for Extensions whose
// referenced clusters have been modified.
func ClusterToExtensionMapper(client client.Client, predicates ...predicate.Predicate) handler.Mapper {
	return controller.ClusterToObjectMapper(client, func() runtime.Object { return &extensions1alpha1.ExtensionList{} }, predicates)
}

// MapperWithinNamespace returns a mapper that returns requests for Extensions whose
// mapper object has been modified in the same namespace.
func MapperWithinNamespace(client client.Client, predicates ...predicate.Predicate) handler.Mapper {
	return controller.MapperWithinNamespace(client, func() runtime.Object { return &extensions1alpha1.ExtensionList{} }, predicates)
}
