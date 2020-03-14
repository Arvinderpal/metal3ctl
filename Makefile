# Copyright 2018 The Kubernetes Authors.
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

BMO_REPO ?= $(HOME)/go/src/github.com/metal3-io/baremetal-operator

install_requirements:
	./hack/install_packages_ubuntu.sh

start_mgmt_cluster:
	./hack/minikube.sh

copy_ironic_bmo_configmap_file:
	cp ./_artifacts/ironic_bmo_configmap.env $(BMO_REPO)/deploy/ironic-keepalived-config/ironic_bmo_configmap.env

delete_mgmt_cluster:
	sudo minikube delete