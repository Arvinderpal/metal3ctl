#!/bin/bash

if ! command -v minikube 2>/dev/null ; then
    curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
    chmod +x minikube
    sudo mv minikube /usr/local/bin/.
fi

if ! command -v docker-machine-driver-kvm2 2>/dev/null ; then
    curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-kvm2
    chmod +x docker-machine-driver-kvm2
    sudo mv docker-machine-driver-kvm2 /usr/local/bin/.
fi

sudo apt-get install -y crudini curl dnsmasq nmap ovmf patch psmisc wget libvirt-bin libvirt-clients libvirt-dev jq nodejs unzip yarn genisoimage qemu-kvm libguestfs-tools gir1.2-polkit-1.0 libpolkit-agent-1-0 libpolkit-backend-1-0 libpolkit-gobject-1-0