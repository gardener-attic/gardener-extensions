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
VERSION          := $(shell cat $(REPO_ROOT)/VERSION)
LD_FLAGS         := "-w -X github.com/gardener/gardener-extensions/gardener-extension-os-coreos/pkg/version.Version=$(IMAGE_TAG)"
VERIFY           := true

### Build commands

.PHONY: format
format:
	@go fmt ./pkg/... ./controllers/...

.PHONY: generate
generate:
	@go generate ./pkg/... ./controllers/...

.PHONY: check
check:
	@./hack/check.sh

.PHONY: test
test:
	@ginkgo -r controllers pkg

.PHONY: verify
verify: check test

.PHONY: install
install:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go install -ldflags $(LD_FLAGS) \
		./controllers/...


.PHONY: all
ifeq ($(VERIFY),true)
all: generate format verify install
else
all: generate format install
endif

### Docker commands

.PHONY: docker-login
docker-login:
	@gcloud auth activate-service-account --key-file .kube-secrets/gcr/gcr-readwrite.json

.PHONY: docker-image-hyper
docker-image-hyper:
	@docker build --build-arg VERIFY=$(VERIFY) -t $(IMAGE_PREFIX)/gardener-extension-hyper:$(VERSION) -t $(IMAGE_PREFIX)/gardener-extension-hyper:latest -f Dockerfile --target gardener-extension-hyper .

.PHONY: docker-image-os-coreos
docker-image-os-coreos:
	@docker build --build-arg VERIFY=$(VERIFY) -t $(IMAGE_PREFIX)/gardener-extension-os-coreos:$(VERSION) -t $(IMAGE_PREFIX)/gardener-extension-os-coreos:latest -f Dockerfile --target gardener-extension-os-coreos .

.PHONY: docker-image-os-coreos-alicloud
docker-image-os-coreos-alicloud:
	@docker build --build-arg VERIFY=$(VERIFY) -t $(IMAGE_PREFIX)/gardener-extension-os-coreos-alicloud:$(VERSION) -t $(IMAGE_PREFIX)/gardener-extension-os-coreos-alicloud:$(VERSION) -f Dockerfile --target gardener-extension-os-coreos-alicloud .


.PHONY: docker-images
docker-images: docker-image-hyper docker-image-os-coreos docker-image-os-coreos-alicloud

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
