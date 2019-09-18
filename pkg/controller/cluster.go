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

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Cluster contains the decoded resources of Gardener's extension Cluster resource.
// TODO: Remove `gardenv1beta1` after one release.
type Cluster struct {
	CloudProfile *gardenv1beta1.CloudProfile
	Seed         *gardenv1beta1.Seed
	Shoot        *gardenv1beta1.Shoot

	CoreCloudProfile *gardencorev1alpha1.CloudProfile
	CoreSeed         *gardencorev1alpha1.Seed
	CoreShoot        *gardencorev1alpha1.Shoot
}

// GetCluster tries to read Gardener's Cluster extension resource in the given namespace.
func GetCluster(ctx context.Context, c client.Client, namespace string) (*Cluster, error) {
	cluster := &extensionsv1alpha1.Cluster{}
	if err := c.Get(ctx, kutil.Key(namespace), cluster); err != nil {
		return nil, err
	}

	decoder, err := NewGardenDecoder()
	if err != nil {
		return nil, err
	}

	cloudProfile, err := CloudProfileFromCluster(decoder, cluster)
	if err != nil {
		return nil, err
	}
	seed, err := SeedFromCluster(decoder, cluster)
	if err != nil {
		return nil, err
	}
	shoot, err := ShootFromCluster(decoder, cluster)
	if err != nil {
		return nil, err
	}

	coreCloudProfile, err := CoreCloudProfileFromCluster(decoder, cluster)
	if err != nil {
		return nil, err
	}
	coreSeed, err := CoreSeedFromCluster(decoder, cluster)
	if err != nil {
		return nil, err
	}
	coreShoot, err := CoreShootFromCluster(decoder, cluster)
	if err != nil {
		return nil, err
	}

	return &Cluster{cloudProfile, seed, shoot, coreCloudProfile, coreSeed, coreShoot}, nil
}

// CloudProfileFromCluster returns the CloudProfile resource inside the Cluster resource.
func CloudProfileFromCluster(decoder runtime.Decoder, cluster *extensionsv1alpha1.Cluster) (*gardenv1beta1.CloudProfile, error) {
	cloudProfile := &gardenv1beta1.CloudProfile{}

	if cluster.Spec.CloudProfile.Raw == nil {
		return nil, nil
	}
	if _, _, err := decoder.Decode(cluster.Spec.CloudProfile.Raw, nil, cloudProfile); err != nil {
		// If cluster.Spec.CloudProfile.Raw is not of type gardenv1beta1.CloudProfile then it is probably
		// of type gardencorev1alpha1.CloudProfile. We don't want to return an error in this case.
		return nil, nil
	}

	return cloudProfile, nil
}

// SeedFromCluster returns the Seed resource inside the Cluster resource.
func SeedFromCluster(decoder runtime.Decoder, cluster *extensionsv1alpha1.Cluster) (*gardenv1beta1.Seed, error) {
	seed := &gardenv1beta1.Seed{}

	if cluster.Spec.Seed.Raw == nil {
		return nil, nil
	}
	if _, _, err := decoder.Decode(cluster.Spec.Seed.Raw, nil, seed); err != nil {
		// If cluster.Spec.Seed.Raw is not of type gardenv1beta1.Seed then it is probably
		// of type gardencorev1alpha1.Seed. We don't want to return an error in this case.
		return nil, nil
	}

	return seed, nil
}

// ShootFromCluster returns the Shoot resource inside the Cluster resource.
func ShootFromCluster(decoder runtime.Decoder, cluster *extensionsv1alpha1.Cluster) (*gardenv1beta1.Shoot, error) {
	shoot := &gardenv1beta1.Shoot{}

	if cluster.Spec.Shoot.Raw == nil {
		return nil, nil
	}
	if _, _, err := decoder.Decode(cluster.Spec.Shoot.Raw, nil, shoot); err != nil {
		// If cluster.Spec.Shoot.Raw is not of type gardenv1beta1.Shoot then it is probably
		// of type gardencorev1alpha1.Shoot. We don't want to return an error in this case.
		return nil, nil
	}

	return shoot, nil
}

// CoreCloudProfileFromCluster returns the CloudProfile resource inside the Cluster resource.
func CoreCloudProfileFromCluster(decoder runtime.Decoder, cluster *extensionsv1alpha1.Cluster) (*gardencorev1alpha1.CloudProfile, error) {
	cloudProfile := &gardencorev1alpha1.CloudProfile{}

	if cluster.Spec.CloudProfile.Raw == nil {
		return nil, nil
	}
	if _, _, err := decoder.Decode(cluster.Spec.CloudProfile.Raw, nil, cloudProfile); err != nil {
		// If cluster.Spec.CloudProfile.Raw is not of type gardencorev1alpha1.CloudProfile then it is probably
		// of type gardenv1beta1.CloudProfile. We don't want to return an error in this case.
		return nil, nil
	}

	return cloudProfile, nil
}

// CoreSeedFromCluster returns the Seed resource inside the Cluster resource.
func CoreSeedFromCluster(decoder runtime.Decoder, cluster *extensionsv1alpha1.Cluster) (*gardencorev1alpha1.Seed, error) {
	seed := &gardencorev1alpha1.Seed{}

	if cluster.Spec.Seed.Raw == nil {
		return nil, nil
	}
	if _, _, err := decoder.Decode(cluster.Spec.Seed.Raw, nil, seed); err != nil {
		// If cluster.Spec.Seed.Raw is not of type gardencorev1alpha1.Seed then it is probably
		// of type gardenv1beta1.Seed. We don't want to return an error in this case.
		return nil, nil
	}

	return seed, nil
}

// CoreShootFromCluster returns the Shoot resource inside the Cluster resource.
func CoreShootFromCluster(decoder runtime.Decoder, cluster *extensionsv1alpha1.Cluster) (*gardencorev1alpha1.Shoot, error) {
	shoot := &gardencorev1alpha1.Shoot{}

	if cluster.Spec.Shoot.Raw == nil {
		return nil, nil
	}
	if _, _, err := decoder.Decode(cluster.Spec.Shoot.Raw, nil, shoot); err != nil {
		// If cluster.Spec.Shoot.Raw is not of type gardencorev1alpha1.Shoot then it is probably
		// of type gardenv1beta1.Shoot. We don't want to return an error in this case.
		return nil, nil
	}

	return shoot, nil
}

// NewGardenDecoder returns a new Garden API decoder.
func NewGardenDecoder() (runtime.Decoder, error) {
	scheme := runtime.NewScheme()
	if err := gardenv1beta1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := gardencorev1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return serializer.NewCodecFactory(scheme).UniversalDecoder(), nil
}
