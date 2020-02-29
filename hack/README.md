
# minikube mgmt cluster

Ensure the required packages are installed on the ubuntu host:

	./install_packages_ubuntu.sh
	
Launch the mgmt cluster:

	minikube.sh


# Useful commands

libvirt/minikube:

	sudo virsh list --all
	sudo virsh domiflist minikube
	sudo minikube ssh ifconfig

Cluster-API:
	

Baremetal-Operator:
	
	kubectl get baremetalhost -n metal3 -oyaml node-0