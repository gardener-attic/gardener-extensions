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

package genericactuator

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	machineGroup   = "machine.sapcloud.io"
	machineVersion = "v1alpha1"
)

var machineCRDs []*apiextensionsv1beta1.CustomResourceDefinition

func init() {
	machineCRDs = []*apiextensionsv1beta1.CustomResourceDefinition{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machinedeployments.machine.sapcloud.io",
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   machineGroup,
				Version: machineVersion,
				Scope:   apiextensionsv1beta1.NamespaceScoped,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Kind:       "MachineDeployment",
					Plural:     "machinedeployments",
					Singular:   "machinedeployment",
					ShortNames: []string{"machdeploy"},
				},
				Subresources: &apiextensionsv1beta1.CustomResourceSubresources{
					Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machinesets.machine.sapcloud.io",
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   machineGroup,
				Version: machineVersion,
				Scope:   apiextensionsv1beta1.NamespaceScoped,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Kind:       "MachineSet",
					Plural:     "machinesets",
					Singular:   "machineset",
					ShortNames: []string{"machset"},
				},
				Subresources: &apiextensionsv1beta1.CustomResourceSubresources{
					Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machines.machine.sapcloud.io",
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   machineGroup,
				Version: machineVersion,
				Scope:   apiextensionsv1beta1.NamespaceScoped,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Kind:       "Machine",
					Plural:     "machines",
					Singular:   "machine",
					ShortNames: []string{"mach"},
				},
				Subresources: &apiextensionsv1beta1.CustomResourceSubresources{
					Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
				},
			},
		},
	}

	machineClasses := map[string]string{
		"alicloud":  "Alicloud",
		"aws":       "AWS",
		"azure":     "Azure",
		"gcp":       "GCP",
		"openstack": "OpenStack",
		"packet":    "Packet",
	}

	for name, kind := range machineClasses {
		machineCRDs = append(machineCRDs, &apiextensionsv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: name + "machineclasses.machine.sapcloud.io",
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   machineGroup,
				Version: machineVersion,
				Scope:   apiextensionsv1beta1.NamespaceScoped,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Kind:       kind + "MachineClass",
					Plural:     name + "machineclasses",
					Singular:   name + "machineclass",
					ShortNames: []string{name + "cls"},
				},
				Subresources: &apiextensionsv1beta1.CustomResourceSubresources{
					Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
				},
			},
		})
	}
}

// TODO: Use github.com/gardener/gardener/pkg/utils/flow.Parallel as soon as we can vendor a new Gardener version again.
func ensureMachineResources(ctx context.Context, c client.Client) error {
	for _, crd := range machineCRDs {
		if err := c.Create(ctx, crd); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}
