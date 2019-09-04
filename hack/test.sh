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


DIRNAME="$(echo "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"
source "$DIRNAME/common.sh"

# Network policies tests must only be ran separately.
SKIP_TESTS=$(echo controllers/provider-{alicloud,aws,azure,gcp,openstack,packet}/test/e2e/networkpolicies | sed 's/ /,/g')

header_text "Test"

if [ $# -gt 0 ]; then
GO111MODULE=on ginkgo -mod=vendor --skipPackage="${SKIP_TESTS}" -r "${@}"
else
GO111MODULE=on ginkgo -mod=vendor --skipPackage="${SKIP_TESTS}" -r "${SOURCE_TREES[@]}"
fi