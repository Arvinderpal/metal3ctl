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
source logging.sh
source common.sh
source network.sh

function configure_minikube() {
    minikube config set vm-driver kvm2
    minikube config set memory 4096
}

#
# Create a mgmt cluster (minikube VM)
#
function init_minikube() {

    #If the vm exists, it has already been initialized
    if [[ "$(sudo virsh list --all)" != *"minikube"* ]]; then
      sudo su -l -c "minikube start" "$USER"
      if [ "${PRE_PULL_IMAGES}" == "true" ]; then
        # Pre-pull the image to reduce pod initialization time
        for IMAGE_VAR in IRONIC_IMAGE IPA_DOWNLOADER_IMAGE IRONIC_INSPECTOR_IMAGE BAREMETAL_OPERATOR_IMAGE; do
          IMAGE=${!IMAGE_VAR}
          sudo su -l -c "minikube ssh sudo docker pull $IMAGE" "${USER}"
        done
      fi
      sudo su -l -c "minikube ssh sudo docker image ls" "${USER}"
      sudo su -l -c "minikube stop" "$USER"
    fi

    # Add network interfaces to mgmt cluster (minikube VM)
    MINIKUBE_IFACES="$(sudo virsh domiflist minikube)"
    # The interface doesn't appear in the minikube VM with --live,
    # so just attach it before next boot. As long as the
    if ! echo "$MINIKUBE_IFACES" | grep -w provisioning  > /dev/null ; then
      sudo virsh attach-interface --domain minikube \
          --model virtio --source provisioning \
          --type network --config
    fi

    if ! echo "$MINIKUBE_IFACES" | grep -w baremetal  > /dev/null ; then
      sudo virsh attach-interface --domain minikube \
          --model virtio --source baremetal \
          --type network --config
    fi
}

function start_minikube_mgmt_cluster(){
  sudo su -l -c 'minikube start' "${USER}"
  # 
  if ! minikube ssh sudo brctl show | grep -w "${CLUSTER_PROVISIONING_INTERFACE}"  > /dev/null ; then
    sudo su -l -c "minikube ssh sudo brctl addbr $CLUSTER_PROVISIONING_INTERFACE" "${USER}"
    sudo su -l -c "minikube ssh sudo ip link set $CLUSTER_PROVISIONING_INTERFACE up" "${USER}"
    sudo su -l -c "minikube ssh sudo ip addr add $INITIAL_IRONICBRIDGE_IP/$PROVISIONING_CIDR dev $CLUSTER_PROVISIONING_INTERFACE" "${USER}"
  fi
  if ! minikube ssh sudo brctl show "${CLUSTER_PROVISIONING_INTERFACE}" | grep -w "eth2"  > /dev/null ; then
    sudo su -l -c "minikube ssh sudo brctl addif $CLUSTER_PROVISIONING_INTERFACE eth2" "${USER}"
    # NOTE(awander): in metal3-dev-env, the following is only executed when IPV6 is enabled. Not sure how that env works w/o this...
    # sudo su -l -c 'minikube ssh "sudo ip addr add '"172.22.0.2/24"' dev eth2"' awander
    sudo su -l -c 'minikube ssh "sudo ip addr add '"$CLUSTER_PROVISIONING_IP/$PROVISIONING_CIDR"' dev eth2"' "${USER}"
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

#
# Write out a clouds.yaml for this environment
#
function create_clouds_yaml() {
  sed -e "s/__CLUSTER_URL_HOST__/$CLUSTER_URL_HOST/g" "${REPO_ROOT}"/hack/clouds.yaml.template > "${REPO_ROOT}"/hack/clouds.yaml
  # To bind this into the ironic-client container we need a directory
  mkdir -p "${REPO_ROOT}"/hack/_clouds_yaml
  cp clouds.yaml "${REPO_ROOT}"/hack/_clouds_yaml/
}

function copy_ironic_bmo_configmap_file() {
  cp "${REPO_ROOT}"/_artifacts/ironic_bmo_configmap.env "${BMO_REPO}/deploy/ironic-keepalived-config/ironic_bmo_configmap.env"
}

create_clouds_yaml
ironic_bmo_configmap_file "${ARTIFACTS}"
configure_minikube
init_minikube
start_minikube_mgmt_cluster
copy_ironic_bmo_configmap_file
