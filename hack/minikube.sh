#!/bin/bash

USER="$(whoami)"
export USER=${USER}

export IRONIC_IMAGE=${IRONIC_IMAGE:-"quay.io/metal3-io/ironic"}
export IPA_DOWNLOADER_IMAGE=${IPA_DOWNLOADER_IMAGE:-"quay.io/metal3-io/ironic-ipa-downloader"}
export IRONIC_INSPECTOR_IMAGE=${IRONIC_INSPECTOR_IMAGE:-"quay.io/metal3-io/ironic-inspector"}
export BAREMETAL_OPERATOR_IMAGE=${BAREMETAL_OPERATOR_IMAGE:-"quay.io/metal3-io/baremetal-operator"}

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

configure_minikube
init_minikube
