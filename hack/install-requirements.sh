#!/bin/bash
#
# Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -e

echo "Installing requirements"
go get -u "gopkg.in/gobuffalo/packr.v1/v2/packr2"
go get -u "gopkg.in/onsi/ginkgo.v1/ginkgo"
go get -u "gopkg.in/golang/mock.v1/mockgen"
go get -u "golang.org/x/lint/golint"
curl -s "https://raw.githubusercontent.com/helm/helm/v2.12.3/scripts/get" | bash
