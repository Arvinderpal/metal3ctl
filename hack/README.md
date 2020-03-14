
# minikube mgmt cluster

Ensure the required packages are installed on the ubuntu host:

	cd hack/
	./install_packages_ubuntu.sh
	
Launch the mgmt cluster:

	PRE_PULL_IMAGES=true ./minikube.sh

# Init BMO on the mgmt cluster
Copy over the generated ConfigMap values to your local baremetal-operator repository. The values will be incorporated into the final ConfigMap used during the deployment: 

	make copy_ironic_bmo_configmap_file

Using the provided example metal3ctl config file, initialize the mgmt cluster with the baremetal-operator:

	./metal3ctl --config examples/metal3ctl.dev.conf init --skip-capi

# Init CAPI components on the mgmt cluster

	./metal3ctl --config examples/metal3ctl.dev.conf init --skip-bmo

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