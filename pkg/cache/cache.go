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

package cache

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clientgocache "k8s.io/client-go/tools/cache"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"strings"
	"sync"
)

// TypeSet stores a set of types in form of runtime.Object.
type TypeSet struct {
	structured   map[reflect.Type]struct{}
	unstructured map[schema.GroupVersionKind]struct{}
}

// Has checks if the given type is already contained in the set.
func (t *TypeSet) Has(obj runtime.Object) bool {
	if u, ok := obj.(*unstructured.Unstructured); ok {
		_, ok = t.unstructured[u.GroupVersionKind()]
		return ok
	}

	_, ok := t.structured[reflect.TypeOf(obj)]
	return ok
}

// Insert inserts the type into the set.
func (t *TypeSet) Insert(obj runtime.Object) {
	if u, ok := obj.(*unstructured.Unstructured); ok {
		t.unstructured[u.GroupVersionKind()] = struct{}{}
		return
	}

	t.structured[reflect.TypeOf(obj)] = struct{}{}
}

// NewTypeSet instantiates set that stores the types of runtime.Objects.
func NewTypeSet(objs ...runtime.Object) *TypeSet {
	set := &TypeSet{
		structured:   make(map[reflect.Type]struct{}),
		unstructured: make(map[schema.GroupVersionKind]struct{}),
	}

	for _, obj := range objs {
		set.Insert(obj)
	}

	return set
}

// New returns a new cache with the informer sync fixed.
// TODO: Remove after migrating to controller-runtime v0.2.x
func New(config *rest.Config, options cache.Options) (cache.Cache, error) {
	c, err := cache.New(config, options)
	if err != nil {
		return nil, err
	}

	s := options.Scheme
	if s == nil {
		s = scheme.Scheme
	}
	return &cacheFix{
		checked: NewTypeSet(),
		scheme:  s,
		stop:    make(chan struct{}),
		Cache:   c,
	}, nil
}

type cacheFix struct {
	checked *TypeSet
	scheme  *runtime.Scheme
	mu      sync.RWMutex
	started bool
	stop    <-chan struct{}
	cache.Cache
}

func (c *cacheFix) notStartedOrChecked(obj runtime.Object) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.started || c.checked.Has(obj)
}

func (c *cacheFix) checkInformerForObjectHasSynced(obj runtime.Object) error {
	if c.notStartedOrChecked(obj) {
		return nil
	}

	informer, err := c.Cache.GetInformer(obj)
	if err != nil {
		return err
	}

	return c.checkInformerHasSynced(obj, informer)
}

func (c *cacheFix) checkInformerHasSynced(obj runtime.Object, informer clientgocache.SharedIndexInformer) error {
	if c.notStartedOrChecked(obj) {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.checked.Has(obj) {
		return nil
	}
	if !clientgocache.WaitForCacheSync(c.stop, informer.HasSynced) {
		return fmt.Errorf("could not sync informer")
	}
	c.checked.Insert(obj)
	return nil
}

func (c *cacheFix) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	if err := c.checkInformerForObjectHasSynced(obj); err != nil {
		return err
	}

	return c.Cache.Get(ctx, key, obj)
}

func (c *cacheFix) objForList(list runtime.Object) (runtime.Object, error) {
	gvk, err := apiutil.GVKForObject(list, c.scheme)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(gvk.Kind, "List") {
		return nil, fmt.Errorf("non-list type %T (kind %q) passed as output", list, gvk)
	}
	// we need the non-list GVK, so chop off the "List" from the end of the kind
	gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	_, isUnstructured := list.(*unstructured.UnstructuredList)
	if isUnstructured {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		return u, nil
	}

	itemsPtr, err := meta.GetItemsPtr(list)
	if err != nil {
		return nil, err
	}

	// http://knowyourmeme.com/memes/this-is-fine
	elemType := reflect.Indirect(reflect.ValueOf(itemsPtr)).Type().Elem()
	cacheTypeValue := reflect.Zero(reflect.PtrTo(elemType))

	obj, ok := cacheTypeValue.Interface().(runtime.Object)
	if !ok {
		return nil, fmt.Errorf("cannot get cache for %T, its element %T is not a runtime.Object", list, cacheTypeValue.Interface())
	}
	return obj, nil
}

func (c *cacheFix) List(ctx context.Context, opts *client.ListOptions, list runtime.Object) error {
	obj, err := c.objForList(list)
	if err != nil {
		return err
	}

	if err := c.checkInformerForObjectHasSynced(obj); err != nil {
		return err
	}

	return c.Cache.List(ctx, opts, list)
}

func (c *cacheFix) GetInformer(obj runtime.Object) (clientgocache.SharedIndexInformer, error) {
	informer, err := c.Cache.GetInformer(obj)
	if err != nil {
		return nil, err
	}

	if err := c.checkInformerHasSynced(obj, informer); err != nil {
		return nil, err
	}

	return informer, nil
}

func (c *cacheFix) GetInformerForKind(gvk schema.GroupVersionKind) (clientgocache.SharedIndexInformer, error) {
	informer, err := c.Cache.GetInformerForKind(gvk)
	if err != nil {
		return nil, err
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	if err := c.checkInformerHasSynced(obj, informer); err != nil {
		return nil, err
	}
	return informer, nil
}

func (c *cacheFix) Start(stopCh <-chan struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.Cache.Start(stopCh); err != nil {
		return err
	}

	c.stop = stopCh
	c.started = true
	return nil
}

func (c *cacheFix) WaitForCacheSync(stop <-chan struct{}) bool {
	return c.Cache.WaitForCacheSync(stop)
}

func (c *cacheFix) IndexField(obj runtime.Object, field string, extractValue client.IndexerFunc) error {
	return c.Cache.IndexField(obj, field, extractValue)
}
