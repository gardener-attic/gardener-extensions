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

package generator

import (
	"github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig/oscommon/generator/test"
	"github.com/gobuffalo/packr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("JeOS Cloud-init Generator Test", func() {
	var box = packr.NewBox("./testfiles/cloud-init")
	os.Setenv(BootCommand, "cloud-init-command")
	os.Setenv(OsConfigFormat, "cloud-init")
	generator, err := NewCloudInitGenerator()

	It("should not fail creating generator", func() {
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Conformance Tests Cloud Init", test.DescribeTest(generator, box))
})
