
# minikube mgmt cluster

Ensure the required packages are installed on the ubuntu host:

	cd hack/
	./install_packages_ubuntu.sh
	
Launch the mgmt cluster:

	PRE_PULL_IMAGES=true ./minikube.sh

# Init BMO on the mgmt cluster
Copy over the generated ConfigMap values to your local baremetal-operator repository. The values will be incorporated into the final ConfigMap used during the deployment: 

	make copy_ironic_bmo_configmap_file

Using the provided example metal3ctl config file, initialize the mgmt cluster with the baremetal-operator and cluster-api components:

	./metal3ctl --config examples/metal3ctl.dev.conf init

Alterntively, skip either bmo or cluster-api initalization:

	./metal3ctl --config examples/metal3ctl.dev.conf init --skip-capi
	./metal3ctl --config examples/metal3ctl.dev.conf init --skip-bmo	

# Create BMH 

The current approach relies on the `02_configure_host.sh` script from the metal3-dev-env repo to configure a couple of VMs that will function as the baremetal hosts:

	export IMAGE_OS=Ubuntu
	export DEFAULT_HOSTS_MEMORY=4096
	export BMOPATH=/home/awander/go/src/github.com/metal3-io/baremetal-operator
	cd $BMOPATH
	./02_configure_host.sh

Create and apply BareMetalHost definitions:
	
	DO_KUBECTL_APPLY=1 ./bmh_create.sh 

# Create Cluster

	kc apply -f hack/capi/v1alpha3/cluster.yaml

# Create first Control-Plane Node

	kc apply -f hack/capi/v1alpha3/control_plane.yaml

# Delete BMO and CAPI components, CRDs, namspaces, etc.

	./metal3ctl --config examples/metal3ctl.dev.conf delete --skip-bmo
	./metal3ctl --config examples/metal3ctl.dev.conf delete --skip-capi

# Delete Minikube cluster

	sudo minikube delete

# Useful commands

libvirt/minikube:

	sudo virsh list --all
	sudo virsh domiflist minikube
	sudo minikube ssh ifconfig

Cluster-API:
	

Baremetal-Operator:
	
	kubectl get baremetalhost -n metal3 
	kubectl describe baremetalhost -n metal3 node-0