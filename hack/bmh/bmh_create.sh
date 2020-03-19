#!/bin/bash

NODES_FILE=$1
BMH_CRS_FILE=$2

function list_nodes() {
    # shellcheck disable=SC2002
    cat "$NODES_FILE" | \
        jq '.nodes[] | {
           name,
           driver,
           address:.driver_info.address,
           port:.driver_info.port,
           user:.driver_info.username,
           password:.driver_info.password,
           mac: .ports[0].address
           } |
           .name + " " +
           .address + " " +
           .user + " " + .password + " " + .mac' \
       | sed 's/"//g'
}

function make_bm_hosts() {
    while read -r name address user password mac; do
        make-bm-worker \
           -address "$address" \
           -password "$password" \
           -user "$user" \
           -boot-mac "$mac" \
           "$name"
    done
}

if [ "${BMOPATH}" == "" ]; then
  echo "Please set env var BMOPATH to your local baremetal-repository."
  exit 1
fi

if [ "$1" == "" ]; then
  NODES_FILE="/opt/metal3-dev-env/ironic_nodes.json"
  echo "Reading from default NODES_FILE file: ${NODES_FILE}"
else
    echo "Reading from NODES_FILE: ${NODES_FILE}"
fi

if [ "$2" == "" ]; then
  BMH_CRS_FILE="bmhosts_crs.yaml"
  echo "Writing to default NODES_FILE file: ${BMH_CRS_FILE}"
else
  echo "Writing to BMH_CRS_FILE: ${BMH_CRS_FILE}"
fi

list_nodes | make_bm_hosts > $BMH_CRS_FILE

if [ -n "${DO_KUBECTL_APPLY}" ]; then
  kubectl apply -f $BMH_CRS_FILE -n metal3
fi