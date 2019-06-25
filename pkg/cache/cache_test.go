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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}

var _ = Describe("Cache", func() {
	Context("TypeSet", func() {
		var (
			structuredType   corev1.ConfigMap
			unstructuredType unstructured.Unstructured
		)
		BeforeEach(func() {
			structuredType = corev1.ConfigMap{}
			unstructuredType = unstructured.Unstructured{}
			unstructuredType.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))
		})

		DescribeTable("#Has",
			func(t runtime.Object) {
				Expect(NewTypeSet().Has(t)).To(BeFalse())
			},
			Entry("structured", &structuredType),
			Entry("unstructured", &unstructuredType))

		DescribeTable("#Insert",
			func(t runtime.Object) {
				ts := NewTypeSet()
				Expect(ts.Has(t)).To(BeFalse())
				ts.Insert(t)
				Expect(ts.Has(t)).To(BeTrue())
			},
			Entry("structured", &structuredType),
			Entry("unstructured", &unstructuredType))
	})
})
