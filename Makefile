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

REGISTRY         := eu.gcr.io/gardener-project
IMAGE_PREFIX     := $(REGISTRY)/gardener-extension-os-coreos
REPO_ROOT        := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR         := $(REPO_ROOT)/hack
VERSION          := $(shell bash -c 'source $(HACK_DIR)/common.sh && echo $$VERSION')
LD_FLAGS         := "-w -X github.com/gardener/gardener-extensions/gardener-extension-os-coreos/pkg/version.Version=$(IMAGE_TAG)"
VERIFY           := true

### Build commands

.PHONY: format
format:
	@./hack/format.sh

.PHONY: clean
clean:
	@./hack/clean.sh

.PHONY: generate
generate:
	@./hack/generate.sh

.PHONY: check
check:
	@./hack/check.sh

.PHONY: test
test:
	@./hack/test.sh

.PHONY: verify
verify: check generate test format

.PHONY: install
install:
	@./hack/install.sh


.PHONY: all
ifeq ($(VERIFY),true)
all: verify generate install
else
all: generate install
endif

### Docker commands

.PHONY: docker-login
docker-login:
	@gcloud auth activate-service-account --key-file .kube-secrets/gcr/gcr-readwrite.json

.PHONY: docker-image-hyper
docker-image-hyper:
	@docker build --build-arg VERIFY=$(VERIFY) -t $(IMAGE_PREFIX)/gardener-extension-hyper:$(VERSION) -t $(IMAGE_PREFIX)/gardener-extension-hyper:latest -f Dockerfile --target gardener-extension-hyper .

.PHONY: docker-images
docker-images: docker-image-hyper

### Debug / Development commands

.PHONY: revendor
revendor:
	@dep ensure -update

.PHONY: start-os-coreos
start-os-coreos:
	@LEADER_ELECTION_NAMESPACE=garden go run -ldflags $(LD_FLAGS) ./controllers/os-coreos/cmd/gardener-extension-os-coreos

.PHONY: start-os-coreos-alicloud
start-os-coreos-alicloud:
	@LEADER_ELECTION_NAMESPACE=garden go run -ldflags $(LD_FLAGS) ./controllers/os-coreos-alicloud/cmd/gardener-extension-os-coreos-alicloud
