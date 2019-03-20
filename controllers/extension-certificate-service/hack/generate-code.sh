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

function headers() {
  echo '''/*
Copyright (c) YEAR SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
'''
}

function usage() {
  cat <<EOM
  Usage: $(basename $0) <output-package> <internal-apis-package> <extensiona-apis-package> <groups-versions>

  <output-package>    the output package name (e.g. github.com/example/project/pkg/generated).
  <ext-apis-package>  the external types dir (e.g. github.com/example/project/pkg/apis or githubcom/example/apis).
  <groups-versions>   the groups and their versions in the format "groupA:v1,v2 groupB:v1 groupC:v2", relative
                      to <api-package>.
EOM
  exit 0
}

OUTPUT_PKG="$1"
EXT_APIS_PKG="$2"
GROUPS_WITH_VERSIONS="$3"

([[ -z "$OUTPUT_PKG" ]] || [[ -z "$EXT_APIS_PKG" ]] || [[ -z "$GROUPS_WITH_VERSIONS" ]]) && usage

rm -f $GOPATH/bin/*-gen

$(dirname $0)/../../../vendor/k8s.io/code-generator/generate-internal-groups.sh \
  deepcopy,defaulter \
  $OUTPUT_PKG \
  $EXT_APIS_PKG \
  $EXT_APIS_PKG \
  $GROUPS_WITH_VERSIONS \
  -h <(headers)

$(dirname $0)/../../../vendor/k8s.io/code-generator/generate-internal-groups.sh \
  conversion \
  $OUTPUT_PKG \
  $EXT_APIS_PKG \
  $EXT_APIS_PKG \
  $GROUPS_WITH_VERSIONS \
  -h <(headers)
