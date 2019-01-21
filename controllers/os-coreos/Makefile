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
IMAGE_REPOSITORY := $(REGISTRY)/gardener-extension-os-coreos
IMAGE_TAG        := $(shell cat ../../VERSION)
WORKDIR          := $(shell pwd)
PUSH_LATEST      := true
LD_FLAGS         := "-w -X github.com/gardener/gardener-extensions/gardener-extension-os-coreos/pkg/version.Version=$(IMAGE_TAG)"

#########################################
# Rules for local development scenarios #
#########################################

.PHONY: start
start:
	@LEADER_ELECTION_NAMESPACE=garden go run \
	-ldflags $(LD_FLAGS) \
		cmd/gardener-extension-os-coreos/*.go \
		--kubeconfig=dev/kubeconfig

#################################################################
# Rules related to binary build, Docker image build and release #
#################################################################

.PHONY: revendor
revendor:
	@dep ensure -update

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags $(LD_FLAGS) \
		-o bin/gardener-extension-os-coreos \
		cmd/gardener-extension-os-coreos/*.go

.PHONY: build-local
build-local:
	@GOBIN=${WORKDIR}/bin go install \
		-ldflags $(LD_FLAGS) \
		./cmd/...

.PHONY: docker-image
docker-image:
	@docker build -t $(IMAGE_REPOSITORY):$(IMAGE_TAG) -t $(IMAGE_REPOSITORY):latest -f Dockerfile --target extension .

.PHONY: docker-login
docker-login:
	@gcloud auth activate-service-account --key-file .kube-secrets/gcr/gcr-readwrite.json

.PHONY: docker-push
docker-push:
	@if ! docker images $(IMAGE_REPOSITORY) | awk '{ print $$2 }' | grep -q -F $(IMAGE_TAG); then echo "$(IMAGE_REPOSITORY) version $(IMAGE_TAG) is not yet built. Please run 'make docker-image'"; false; fi
	@gcloud docker -- push $(IMAGE_REPOSITORY):$(IMAGE_TAG)
	@if [[ "$(PUSH_LATEST)" == "true" ]]; then gcloud docker -- push $(IMAGE_REPOSITORY):latest; fi

.PHONY: rename-binaries
rename-binaries:
	@if [[ -f bin/gardener-extension-os-coreos ]]; then cp bin/gardener-extension-os-coreos gardener-extension-os-coreos-darwin-amd64; fi
	@if [[ -f bin/rel/gardener-extension-os-coreos ]]; then cp bin/rel/gardener-extension-os-coreos gardener-extension-os-coreos-linux-amd64; fi

.PHONY: clean
clean:
	@rm -rf bin/
	@rm -f *linux-amd64
	@rm -f *darwin-amd64

#####################################################################
# Rules for verification, formatting, linting, testing and cleaning #
#####################################################################

.PHONY: verify
verify: check test

.PHONY: check
check:
	@.ci/check

.PHONY: test
test:
	@.ci/test
