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

header_text "Check"

echo "Executing check-generate"
"$DIRNAME"/check-generate.sh

echo "Executing golangci-lint"
golangci-lint run -c "$DIRNAME"/../.golangci.yaml "${SOURCE_TREES[@]}"

echo "Checking for format issues with gofmt"
unformatted_files="$(gofmt -l controllers pkg)"
if [[ "$unformatted_files" ]]; then
    echo "Unformatted files detected:"
    echo "$unformatted_files"
    exit 1
fi

echo "Checking for chart symlink errors"
BROKEN_SYMLINKS=$(find -L controllers/*/charts -type l)
if [[ "$BROKEN_SYMLINKS" ]]; then
   echo "Found broken symlinks:"
   echo "$BROKEN_SYMLINKS"
   exit 1
fi

echo "Checking whether all charts can be rendered"
for chart_file in controllers/*/charts/*/Chart.yaml; do
    helm template "$(dirname "$chart_file")" 1> /dev/null
done

echo "All checks successful"
