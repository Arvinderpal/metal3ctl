#!/bin/bash

# Copyright 2020 The Kubernetes Authors.
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

USER="$(whoami)"
export USER=${USER}

export IRONIC_IMAGE=${IRONIC_IMAGE:-"quay.io/metal3-io/ironic"}
export IPA_DOWNLOADER_IMAGE=${IPA_DOWNLOADER_IMAGE:-"quay.io/metal3-io/ironic-ipa-downloader"}
export IRONIC_INSPECTOR_IMAGE=${IRONIC_INSPECTOR_IMAGE:-"quay.io/metal3-io/ironic-inspector"}
export BAREMETAL_OPERATOR_IMAGE=${BAREMETAL_OPERATOR_IMAGE:-"quay.io/metal3-io/baremetal-operator"}

REPO_ROOT=$(git rev-parse --show-toplevel)
export ARTIFACTS="${ARTIFACTS:-${REPO_ROOT}/_artifacts}"
mkdir -p "$ARTIFACTS/"

# Cluster.
export CLUSTER_NAME=${CLUSTER_NAME:-"test1"}
export CLUSTER_APIENDPOINT_IP=${CLUSTER_APIENDPOINT_IP:-"192.168.111.249"}
export KUBERNETES_VERSION=${KUBERNETES_VERSION:-"v1.17.0"}

export SSH_KEY=${SSH_KEY:-"${HOME}/.ssh/id_rsa"}
export SSH_PUB_KEY=${SSH_PUB_KEY:-"${SSH_KEY}.pub"}
# Generate user ssh key
if [ ! -f "${SSH_KEY}" ]; then
  mkdir -p "$(dirname "$SSH_KEY")"
  ssh-keygen -f "${SSH_KEY}" -P ""
fi
export SSH_PRIVATE_KEY=`cat ${SSH_KEY}`
export SSH_PUB_KEY_CONTENT=`cat ${SSH_PUB_KEY}`
