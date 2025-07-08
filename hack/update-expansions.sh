#!/usr/bin/env bash

# Copyright 2025 The KCP Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

CLIENTGO_PKG="$(go list -f '{{.Dir}}' -m k8s.io/client-go)"

list_expansions() {
    find kubernetes/typed -name "*_expansion.go" \
        | grep -v generated_expansion
}

update_expansion() {
    local upstream_file="$1"
    local local_file="$2"
    # TODO could look into making this a go command and then use the AST
    # to make intelligent transformations.
    # E.g. the types are somewhat deterministic from `fakeX` to `scopedX`
    # On the other hand that is a lot of work to save a few seconds once
    # every few months.
    sed \
        -e '/Copyright .* The Kubernetes Authors./a \
Modifications Copyright YEAR The KCP Authors.' \
        -e 's#k8s.io/client-go/testing#github.com/kcp-dev/client-go/third_party/k8s.io/client-go/testing#' \
        "$upstream_file" > "$local_file"
}

# update existing expansions
for expansion in $(list_expansions); do
    update_expansion "$CLIENTGO_PKG/$expansion" "$expansion"
done

# copy any new fake expansions from upstream
for upstream_expansion in $(find "$CLIENTGO_PKG/kubernetes/typed" -name "*_expansion.go" | grep fake); do
    local_equivalent="${upstream_expansion##$CLIENTGO_PKG/}"
    if [[ -f "$local_equivalent" ]]; then
        echo "Skipping $local_equivalent, already exists"
    else
        update_expansion "$upstream_expansion" "$local_equivalent"
    fi
done
