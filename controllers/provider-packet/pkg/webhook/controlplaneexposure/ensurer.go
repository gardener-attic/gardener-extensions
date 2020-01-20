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

package controlplaneexposure

import (
	"context"

	"github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/apis/config"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/genericmutator"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/resource"
)

// NewEnsurer creates a new controlplaneexposure ensurer.
func NewEnsurer(etcdStorage *config.ETCDStorage, logger logr.Logger) genericmutator.Ensurer {
	return &ensurer{
		etcdStorage: etcdStorage,
		logger:      logger.WithName("packet-controlplaneexposure-ensurer"),
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer
	etcdStorage *config.ETCDStorage
	logger      logr.Logger
}

// EnsureETCD ensures that the etcd conform to the provider requirements.
func (e *ensurer) EnsureETCD(ctx context.Context, ectx genericmutator.EnsurerContext, etcd *druidv1alpha1.Etcd) error {
	var (
		etcdStorage config.ETCDStorage
	)

	capacity := resource.MustParse("10Gi")
	class := ""
	if etcd.Name == v1beta1constants.ETCDMain {
		etcdStorage = *e.etcdStorage
		if etcdStorage.Capacity != nil {
			capacity = *etcdStorage.Capacity
		}
		if etcdStorage.ClassName != nil {
			class = *etcdStorage.ClassName
		}
	}
	// Determine the storage capacity
	// A non-default storage capacity is used only if it's configured

	etcd.Spec.StorageClass = &class
	etcd.Spec.StorageCapacity = &capacity
	return nil
}
