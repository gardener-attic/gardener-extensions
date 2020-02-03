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

.PHONY: format
format:
	@./hack/format.sh ./pkg

.PHONY: clean
clean:
	@./hack/clean.sh ./pkg/...

.PHONY: generate
generate:
	@./hack/generate.sh ./...

.PHONY: check
check:
	@./hack/check.sh ./...

.PHONY: test
test:
	@./hack/test.sh -r ./...

.PHONY: verify
verify: check generate test format

.PHONY: install-requirements
install-requirements:
	@go install -mod=vendor ./vendor/github.com/gobuffalo/packr/v2/packr2
	@go install -mod=vendor ./vendor/github.com/golang/mock/mockgen
	@go install -mod=vendor ./vendor/github.com/onsi/ginkgo/ginkgo
	@./hack/install-requirements.sh

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod vendor
	@GO111MODULE=on go mod tidy

.PHONY: all
ifeq ($(VERIFY),true)
all: verify generate install
else
all: generate install
endif
