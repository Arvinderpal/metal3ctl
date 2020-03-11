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
set -xe
# shellcheck disable=SC1091
source common.sh
source network.sh

function configure_minikube() {
    minikube config set vm-driver kvm2
    minikube config set memory 4096
}

function init_minikube() {

    #If the vm exists, it has already been initialized
    if [[ "$(sudo virsh list --all)" != *"minikube"* ]]; then
      sudo su -l -c "minikube start" "$USER"
      # Pre-pull the image to reduce pod initialization time
      for IMAGE_VAR in IRONIC_IMAGE IPA_DOWNLOADER_IMAGE IRONIC_INSPECTOR_IMAGE BAREMETAL_OPERATOR_IMAGE; do
        IMAGE=${!IMAGE_VAR}
        sudo su -l -c "minikube ssh sudo docker pull $IMAGE" "${USER}"
      done
      sudo su -l -c "minikube ssh sudo docker image ls" "${USER}"
      # sudo su -l -c "minikube stop" "$USER"
    fi

}

function ironic_bmo_configmap_file() {
  filepath=$1
cat <<EOF> "$filepath/ironic_bmo_configmap.env"
HTTP_PORT=6180
PROVISIONING_IP=$CLUSTER_PROVISIONING_IP
PROVISIONING_INTERFACE=$CLUSTER_PROVISIONING_INTERFACE
PROVISIONING_CIDR=$PROVISIONING_CIDR
DHCP_RANGE=$CLUSTER_DHCP_RANGE
DEPLOY_KERNEL_URL=http://$CLUSTER_URL_HOST:6180/images/ironic-python-agent.kernel
DEPLOY_RAMDISK_URL=http://$CLUSTER_URL_HOST:6180/images/ironic-python-agent.initramfs
IRONIC_ENDPOINT=http://$CLUSTER_URL_HOST:6385/v1/
IRONIC_INSPECTOR_ENDPOINT=http://$CLUSTER_URL_HOST:5050/v1/
CACHEURL=http://$PROVISIONING_URL_HOST/images
IRONIC_FAST_TRACK=false
EOF
}

ironic_bmo_configmap_file "${ARTIFACTS}"
configure_minikube
init_minikube
