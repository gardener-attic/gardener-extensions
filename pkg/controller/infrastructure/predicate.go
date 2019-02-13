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

package infrastructure

import (
	"strings"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// TypePredicate filters the incoming Infrastructure resources for ones that have the same type
// as the given type.
func TypePredicate(typeName string) predicate.Predicate {
	typeMatches := func(obj runtime.Object) bool {
		if config, ok := obj.(*extensionsv1alpha1.Infrastructure); ok {
			return strings.ToLower(config.Spec.Type) == typeName
		}
		return false
	}

	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return typeMatches(event.Object)
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return typeMatches(updateEvent.ObjectOld)
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return typeMatches(deleteEvent.Object)
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return typeMatches(genericEvent.Object)
		},
	}
}

type generationChangedPredicate struct {
	predicate.Funcs
}

func (generationChangedPredicate) Update(e event.UpdateEvent) bool {
	return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
}

// GenerationChangedPredicate is a predicate for generation changes.
func GenerationChangedPredicate() predicate.Predicate {
	return generationChangedPredicate{}
}
