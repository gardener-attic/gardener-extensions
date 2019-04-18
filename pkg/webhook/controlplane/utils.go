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

package controlplane

import (
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// ContainerWithName returns the container with the given name if it exists in the given slice, nil otherwise.
func ContainerWithName(containers []corev1.Container, name string) *corev1.Container {
	if i := containerWithNameIndex(containers, name); i >= 0 {
		return &containers[i]
	}
	return nil
}

// EnsureStringWithPrefix ensures that a string having the given prefix exists in the given slice
// with a value equal to prefix + value.
func EnsureStringWithPrefix(items []string, prefix, value string) []string {
	item := prefix + value
	if i := StringWithPrefixIndex(items, prefix); i < 0 {
		items = append(items, item)
	} else if items[i] != item {
		items = append(append(items[:i], item), items[i+1:]...)
	}
	return items
}

// EnsureNoStringWithPrefix ensures that a string having the given prefix does not exist in the given slice.
func EnsureNoStringWithPrefix(items []string, prefix string) []string {
	if i := StringWithPrefixIndex(items, prefix); i >= 0 {
		items = append(items[:i], items[i+1:]...)
	}
	return items
}

// EnsureStringWithPrefixContains ensures that a string having the given prefix exists in the given slice
// and contains the given value in a list separated by sep.
func EnsureStringWithPrefixContains(items []string, prefix, value, sep string) []string {
	if i := StringWithPrefixIndex(items, prefix); i < 0 {
		items = append(items, prefix+value)
	} else {
		values := strings.Split(strings.TrimPrefix(items[i], prefix), sep)
		if j := StringIndex(values, value); j < 0 {
			values = append(values, value)
			items = append(append(items[:i], prefix+strings.Join(values, sep)), items[i+1:]...)
		}
	}
	return items
}

// EnsureNoStringWithPrefixContains ensures that either a string having the given prefix does not exist in the given slice,
// or it doesn't contain the given value in a list separated by sep.
func EnsureNoStringWithPrefixContains(items []string, prefix, value, sep string) []string {
	if i := StringWithPrefixIndex(items, prefix); i >= 0 {
		values := strings.Split(strings.TrimPrefix(items[i], prefix), sep)
		if j := StringIndex(values, value); j >= 0 {
			values = append(values[:j], values[j+1:]...)
			items = append(append(items[:i], prefix+strings.Join(values, sep)), items[i+1:]...)
		}
	}
	return items
}

// EnsureEnvVarWithName ensures that a EnvVar with a name equal to the name of the given EnvVar exists
// in the given slice and is equal to the given EnvVar.
func EnsureEnvVarWithName(items []corev1.EnvVar, item corev1.EnvVar) []corev1.EnvVar {
	if i := envVarWithNameIndex(items, item.Name); i < 0 {
		items = append(items, item)
	} else if !reflect.DeepEqual(items[i], item) {
		items = append(append(items[:i], item), items[i+1:]...)
	}
	return items
}

// EnsureNoEnvVarWithName ensures that a EnvVar with the given name does not exist in the given slice.
func EnsureNoEnvVarWithName(items []corev1.EnvVar, name string) []corev1.EnvVar {
	if i := envVarWithNameIndex(items, name); i >= 0 {
		items = append(items[:i], items[i+1:]...)
	}
	return items
}

// EnsureVolumeMountWithName ensures that a VolumeMount with a name equal to the name of the given VolumeMount exists
// in the given slice and is equal to the given VolumeMount.
func EnsureVolumeMountWithName(items []corev1.VolumeMount, item corev1.VolumeMount) []corev1.VolumeMount {
	if i := volumeMountWithNameIndex(items, item.Name); i < 0 {
		items = append(items, item)
	} else if !reflect.DeepEqual(items[i], item) {
		items = append(append(items[:i], item), items[i+1:]...)
	}
	return items
}

// EnsureNoVolumeMountWithName ensures that a VolumeMount with the given name does not exist in the given slice.
func EnsureNoVolumeMountWithName(items []corev1.VolumeMount, name string) []corev1.VolumeMount {
	if i := volumeMountWithNameIndex(items, name); i >= 0 {
		items = append(items[:i], items[i+1:]...)
	}
	return items
}

// EnsureVolumeWithName ensures that a Volume with a name equal to the name of the given Volume exists
// in the given slice and is equal to the given Volume.
func EnsureVolumeWithName(items []corev1.Volume, item corev1.Volume) []corev1.Volume {
	if i := volumeWithNameIndex(items, item.Name); i < 0 {
		items = append(items, item)
	} else if !reflect.DeepEqual(items[i], item) {
		items = append(append(items[:i], item), items[i+1:]...)
	}
	return items
}

// EnsureNoVolumeWithName ensures that a Volume with the given name does not exist in the given slice.
func EnsureNoVolumeWithName(items []corev1.Volume, name string) []corev1.Volume {
	if i := volumeWithNameIndex(items, name); i >= 0 {
		items = append(items[:i], items[i+1:]...)
	}
	return items
}

// StringIndex returns the index of the first occurrence of the given string in the given slice, or -1 if not found.
func StringIndex(items []string, value string) int {
	for i, item := range items {
		if item == value {
			return i
		}
	}
	return -1
}

// StringWithPrefixIndex returns the index of the first occurrence of a string having the given prefix in the given slice, or -1 if not found.
func StringWithPrefixIndex(items []string, prefix string) int {
	for i, item := range items {
		if strings.HasPrefix(item, prefix) {
			return i
		}
	}
	return -1
}

func containerWithNameIndex(items []corev1.Container, name string) int {
	for i, item := range items {
		if item.Name == name {
			return i
		}
	}
	return -1
}

func envVarWithNameIndex(items []corev1.EnvVar, name string) int {
	for i, item := range items {
		if item.Name == name {
			return i
		}
	}
	return -1
}

func volumeMountWithNameIndex(items []corev1.VolumeMount, name string) int {
	for i, item := range items {
		if item.Name == name {
			return i
		}
	}
	return -1
}

func volumeWithNameIndex(items []corev1.Volume, name string) int {
	for i, item := range items {
		if item.Name == name {
			return i
		}
	}
	return -1
}
