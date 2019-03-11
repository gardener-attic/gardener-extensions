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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PredicatesMatch returns true if all predicates match for the given object.
func PredicatesMatch(predicates []predicate.Predicate, obj runtime.Object) bool {
	m, err := meta.Accessor(obj)
	if err != nil {
		return false
	}

	e := event.GenericEvent{
		Meta:   m,
		Object: obj,
	}

	for _, predicate := range predicates {
		if !predicate.Generic(e) {
			return false
		}
	}

	return true
}

// ShootFailedPredicate is a predicate for failed shoots.
func ShootFailedPredicate(c client.Client) predicate.Predicate {
	ctx := context.TODO()

	shootNotFailed := func(obj runtime.Object) bool {
		accessor, err := meta.Accessor(obj)
		if err != nil {
			return false
		}

		cluster, err := GetCluster(ctx, c, accessor.GetNamespace())
		if err != nil {
			return false
		}

		return !ShootIsFailed(cluster.Shoot)
	}

	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return shootNotFailed(event.Object)
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return shootNotFailed(event.ObjectNew)
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return shootNotFailed(event.Object)
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return shootNotFailed(event.Object)
		},
	}
}

var generationChangedPredicate = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
	},
}

// GenerationChangedPredicate is a predicate for generation changes.
func GenerationChangedPredicate() predicate.Predicate {
	return generationChangedPredicate
}

var annotationsChangedPredicate = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		return !equality.Semantic.DeepEqual(e.MetaOld.GetAnnotations(), e.MetaNew.GetAnnotations())
	},
}

// AnnotationsChangedPredicate is a predicate for annotations changes.
func AnnotationsChangedPredicate() predicate.Predicate {
	return annotationsChangedPredicate
}

// OrPredicate is a predicate for annotations changes.
func OrPredicate(predicates ...predicate.Predicate) predicate.Predicate {
	orRange := func(f func(predicate.Predicate) bool) bool {
		for _, p := range predicates {
			if f(p) {
				return true
			}
		}
		return false
	}

	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return orRange(func(p predicate.Predicate) bool { return p.Create(event) })
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return orRange(func(p predicate.Predicate) bool { return p.Update(event) })
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return orRange(func(p predicate.Predicate) bool { return p.Delete(event) })
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return orRange(func(p predicate.Predicate) bool { return p.Generic(event) })
		},
	}
}
