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

REGISTRY                    := eu.gcr.io/gardener-project
IMAGE_PREFIX                := $(REGISTRY)/gardener
REPO_ROOT                   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR                    := $(REPO_ROOT)/hack
HOSTNAME                    := $(shell hostname)
VERSION                     := $(shell bash -c 'source $(HACK_DIR)/common.sh && echo $$VERSION')
LD_FLAGS                    := "-w -X github.com/gardener/gardener-extensions/pkg/version.Version=$(IMAGE_TAG)"
VERIFY                      := true
LEADER_ELECTION             := false
IGNORE_OPERATION_ANNOTATION := false
CERTIFICATE_SERVICE_CONFIG  := ./controllers/extension-certificate-service/example/config.yaml

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
	@docker build --build-arg VERIFY=$(VERIFY) -t $(IMAGE_PREFIX)/gardener-extension-hyper:$(VERSION) -t $(IMAGE_PREFIX)/gardener-extension-hyper:latest -f Dockerfile -m 6g --target gardener-extension-hyper .

.PHONY: docker-images
docker-images: docker-image-hyper

### Debug / Development commands

.PHONY: revendor
revendor:
	@dep ensure -v -update

.PHONY: start-os-coreos
start-os-coreos:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./controllers/os-coreos/cmd/gardener-extension-os-coreos \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION)

.PHONY: start-os-jeos
start-os-jeos:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./controllers/os-suse-jeos/cmd/gardener-extension-os-suse-jeos \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=false

.PHONY: start-os-coreos-alicloud
start-os-coreos-alicloud:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./controllers/os-coreos-alicloud/cmd/gardener-extension-os-coreos-alicloud \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION)

.PHONY: start-provider-aws
start-provider-aws:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./controllers/provider-aws/cmd/gardener-extension-provider-aws \
		--config-file=./controllers/provider-aws/example/00-componentconfig.yaml \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION) \
		--webhook-config-mode=url \
		--webhook-config-name=gardener-extension-provider-aws \
		--webhook-config-host=$(HOSTNAME)

.PHONY: start-provider-azure
start-provider-azure:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./controllers/provider-azure/cmd/gardener-extension-provider-azure \
		--config-file=./controllers/provider-azure/example/00-componentconfig.yaml \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION) \
		--webhook-config-mode=url \
  		--webhook-config-name=gardener-extension-provider-azure \
   		--webhook-config-host=$(HOSTNAME)

.PHONY: start-provider-gcp
start-provider-gcp:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./controllers/provider-gcp/cmd/gardener-extension-provider-gcp \
		--config-file=./controllers/provider-gcp/example/00-componentconfig.yaml \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION) \
		--webhook-config-mode=url \
		--webhook-config-name=gardener-extension-provider-gcp \
		--webhook-config-host=$(HOSTNAME)

.PHONY: start-provider-openstack
start-provider-openstack:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./controllers/provider-openstack/cmd/gardener-extension-provider-openstack \
		--config-file=./controllers/provider-openstack/example/00-componentconfig.yaml \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION) \
		--webhook-config-mode=url \
		--webhook-config-name=gardener-extension-provider-openstack \
		--webhook-config-host=$(HOSTNAME)

.PHONY: start-provider-alicloud
start-provider-alicloud:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./controllers/provider-alicloud/cmd/gardener-extension-provider-alicloud \
		--config-file=./controllers/provider-alicloud/example/00-componentconfig.yaml \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION) \
		--webhook-config-mode=url \
		--webhook-config-name=gardener-extension-provider-alicloud \
		--webhook-config-host=$(HOSTNAME)

.PHONY: start-provider-packet
start-provider-packet:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./controllers/provider-packet/cmd/gardener-extension-provider-packet \
		--config-file=./controllers/provider-packet/example/00-componentconfig.yaml \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION) \
		--webhook-config-mode=url \
		--webhook-config-name=gardener-extension-provider-packet \
		--webhook-config-host=$(HOSTNAME)

.PHONY: start-certificate-service
start-certificate-service:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./controllers/extension-certificate-service/cmd \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION) \
		--config=$(CERTIFICATE_SERVICE_CONFIG)
