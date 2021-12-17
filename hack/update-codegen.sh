#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
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

OUTPUT_BASE_DIR="/Users/alanfok/Desktop"
SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
cd ${SCRIPT_ROOT}

echo "job_controller gen start"

bash /Users/alanfok/Desktop/Projects/public/et-operator-extension/hack/generate-groups.sh \
   "deepcopy,defaulter" \
  github.com/kube-queue/et-operator-extension/pkg/et-operator/client/common \
  github.com/kube-queue/et-operator-extension/pkg/et-operator/apis/common \
  "job_controller:v1" \
  --output-base ${OUTPUT_BASE_DIR} \
  --go-header-file /Users/alanfok/Desktop/Projects/public/et-operator-extension/hack/boilerplate/boilerplate.generatego.txt

echo "et v1alpha1 gen start"

bash /Users/alanfok/Desktop/Projects/public/et-operator-extension/hack/generate-groups.sh \
  all \
  github.com/kube-queue/et-operator-extension/pkg/et-operator/client \
  github.com/kube-queue/et-operator-extension/pkg/et-operator/apis \
  "et:v1alpha1" \
  --output-base ${OUTPUT_BASE_DIR} \
  --go-header-file /Users/alanfok/Desktop/Projects/public/et-operator-extension/hack/boilerplate.go.txt

echo "Generating defaulters for et/v1alpha1"
${GOPATH}/bin/defaulter-gen  --input-dirs github.com/kube-queue/et-operator-extension/pkg/et-operator/apis/et/v1alpha1 -O zz_generated.defaults --go-header-file /Users/alanfok/Desktop/Projects/public/et-operator-extension/hack/boilerplate/boilerplate.generatego.txt "$@"
