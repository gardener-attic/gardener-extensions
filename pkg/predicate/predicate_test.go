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

package predicate_test

import (
	"encoding/json"
	"github.com/gardener/gardener-extensions/pkg/predicate"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("Predicate", func() {
	Describe("#HasType", func() {
		var (
			extensionType string
			object        runtime.Object
			createEvent   event.CreateEvent
			updateEvent   event.UpdateEvent
			deleteEvent   event.DeleteEvent
			genericEvent  event.GenericEvent
		)

		BeforeEach(func() {
			extensionType = "extension-type"
			object = &extensionDummy{
				extensionType: extensionType,
			}
			createEvent = event.CreateEvent{
				Object: object,
			}
			updateEvent = event.UpdateEvent{
				ObjectOld: object,
				ObjectNew: object,
			}
			deleteEvent = event.DeleteEvent{
				Object: object,
			}
			genericEvent = event.GenericEvent{
				Object: object,
			}
		})

		It("should match the type", func() {
			predicate := predicate.HasType(extensionType)

			Expect(predicate.Create(createEvent)).To(BeTrue())
			Expect(predicate.Update(updateEvent)).To(BeTrue())
			Expect(predicate.Delete(deleteEvent)).To(BeTrue())
			Expect(predicate.Generic(genericEvent)).To(BeTrue())
		})

		It("should not match the type", func() {
			predicate := predicate.HasType("anotherType")

			Expect(predicate.Create(createEvent)).To(BeFalse())
			Expect(predicate.Update(updateEvent)).To(BeFalse())
			Expect(predicate.Delete(deleteEvent)).To(BeFalse())
			Expect(predicate.Generic(genericEvent)).To(BeFalse())
		})
	})

	Describe("#HasName", func() {
		var (
			name         string
			createEvent  event.CreateEvent
			updateEvent  event.UpdateEvent
			deleteEvent  event.DeleteEvent
			genericEvent event.GenericEvent
		)

		BeforeEach(func() {
			objectMeta := metav1.ObjectMeta{
				Name: name,
			}
			createEvent = event.CreateEvent{
				Meta: &objectMeta,
			}
			updateEvent = event.UpdateEvent{
				MetaNew: &objectMeta,
				MetaOld: &objectMeta,
			}
			deleteEvent = event.DeleteEvent{
				Meta: &objectMeta,
			}
			genericEvent = event.GenericEvent{
				Meta: &objectMeta,
			}
		})

		It("should match the name", func() {
			predicate := predicate.HasName(name)

			Expect(predicate.Create(createEvent)).To(BeTrue())
			Expect(predicate.Update(updateEvent)).To(BeTrue())
			Expect(predicate.Delete(deleteEvent)).To(BeTrue())
			Expect(predicate.Generic(genericEvent)).To(BeTrue())
		})

		It("should not match the name", func() {
			predicate := predicate.HasName("anotherName")

			Expect(predicate.Create(createEvent)).To(BeFalse())
			Expect(predicate.Update(updateEvent)).To(BeFalse())
			Expect(predicate.Delete(deleteEvent)).To(BeFalse())
			Expect(predicate.Generic(genericEvent)).To(BeFalse())
		})
	})

	DescribeTable("#ClusterCloudProfileGenerationChanged",
		func(oldMachine, newMachine string, conditionMatcher types.GomegaMatcher) {
			oldCloudProfile := &v1beta1.CloudProfile{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CloudProfile",
					APIVersion: "garden.sapcloud.io/v1beta1",
				},
				Spec: v1beta1.CloudProfileSpec{
					AWS: &v1beta1.AWSProfile{
						Constraints: v1beta1.AWSConstraints{
							MachineTypes: []v1beta1.MachineType{
								v1beta1.MachineType{
									Name: oldMachine,
								},
							},
						},
					},
				},
			}
			newCloudProfile := oldCloudProfile.DeepCopy()
			newCloudProfile.Spec.AWS.Constraints.MachineTypes = []v1beta1.MachineType{v1beta1.MachineType{Name: newMachine}}

			updateEvent := event.UpdateEvent{
				ObjectNew: &v1alpha1.Cluster{
					Spec: v1alpha1.ClusterSpec{
						CloudProfile: runtime.RawExtension{
							Raw: encode(newCloudProfile),
						},
					},
				},
				ObjectOld: &v1alpha1.Cluster{
					Spec: v1alpha1.ClusterSpec{
						CloudProfile: runtime.RawExtension{
							Raw: encode(oldCloudProfile),
						},
					},
				},
			}

			Expect(predicate.ClusterCloudProfileGenerationChanged().Update(updateEvent)).To(conditionMatcher)
		},
		Entry("no update", "machineFoo", "machineFoo", BeFalse()),
		Entry("generation update", "machineFoo", "machineBar", BeTrue()),
	)
})

type extensionDummy struct {
	v1alpha1.Extension
	extensionType string
}

func (e *extensionDummy) GetExtensionType() string {
	return e.extensionType
}

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}
