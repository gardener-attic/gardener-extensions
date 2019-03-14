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
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PredicateLog is the logger for predicates.
var PredicateLog logr.Logger = log.Log

// EvalGenericPredicate returns true if all predicates match for the given object.
func EvalGenericPredicate(predicates []predicate.Predicate, obj runtime.Object) bool {
	e := NewGenericEventFromObject(obj)

	for _, p := range predicates {
		if !p.Generic(e) {
			return false
		}
	}

	return true
}

// ShootFailedPredicate is a predicate for failed shoots.
func ShootFailedPredicate(c client.Client) predicate.Predicate {
	ctx := context.TODO()
	log := PredicateLog.WithName("shoot-failed")

	shootNotFailed := func(log logr.Logger, meta metav1.Object) bool {
		cluster, err := GetCluster(ctx, c, meta.GetNamespace())
		if err != nil {
			log.Info("Could not retrieve corresponding cluster", "error", err.Error())
			return false
		}

		return !ShootIsFailed(cluster.Shoot)
	}

	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return shootNotFailed(CreateEventLogger(log, event), event.Meta)
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return shootNotFailed(UpdateEventLogger(log, event), event.MetaNew)
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return shootNotFailed(DeleteEventLogger(log, event), event.Meta)
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return shootNotFailed(GenericEventLogger(log, event), event.Meta)
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
