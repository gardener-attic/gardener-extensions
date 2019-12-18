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

header_text "Generate"

# Temporary trick avoiding the cyclic dependendies between gardener/gardener-extensions and
# gardener/gardener (for backwards compatibility g/g needs to know the extensions APIs, but
# the extensions need to know the gardener APIs as well).
# Will be removed in the future agian once the deprecated garden.sapcloud.io API group is removed.
(
  export GO111MODULE=on
  export GOFLAGS="-mod=vendor"

  for dir in "${SOURCE_TREES[@]}"; do
    if [[ "$(basename "$dir")" != "..." ]]; then
      go generate "$dir"
      continue
    fi

    found=( )
    while IFS= read -r a; do
      if [[ -z "$a" ]]; then
        continue
      fi

      found=( "${found[@]}" "$a" )
      go generate "$a/..."
    done <<< "$(find "$(dirname "$dir")" -type d | grep 'pkg/apis$')"

    find "$(dirname "$dir")" -type d | grep -v 'pkg/apis' | while IFS= read -r a; do
      for f in "${found[@]}"; do
        if [[ "${f##$a}" != "$f" ]]; then
          continue 2
        fi
      done

      if ls "$a"/*.go >/dev/null 2>&1; then
        go generate "$a"
      fi
    done
  done
)
