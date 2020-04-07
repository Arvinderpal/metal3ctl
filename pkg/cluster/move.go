/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Arvinderpal/metal3ctl/config"
	"github.com/Arvinderpal/metal3ctl/pkg/internal/proxy"
	"github.com/Arvinderpal/metal3ctl/pkg/internal/util"
	"github.com/pkg/errors"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	logf "sigs.k8s.io/cluster-api/cmd/clusterctl/log"
	"sigs.k8s.io/cluster-api/controllers/noderefutil"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bmh "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"
	capm3 "github.com/metal3-io/cluster-api-provider-metal3/api/v1alpha3"
)

type MoveOptions struct {
	FromKubeconfig string
	Namespace      string
	ToKubeconfig   string
	SkipBMO        bool
	SkipCAPI       bool
}

func MoveFromBootstrapToTargetCluster(input config.LoadMetal3CtlConfigInput, options *MoveOptions) error {
	log := logf.Log
	ctx := context.TODO()
	config, err := config.LoadMetal3CtlConfig(ctx, input)
	if err != nil {
		return errors.Wrapf(err, "error loading metal3ctl config file")
	}

	pFrom := proxy.NewProxy(options.FromKubeconfig)
	cFrom, err := pFrom.NewClient()
	if err != nil {
		return errors.Wrap(err, "failed to create controller-runtime client")
	}
	pTo := proxy.NewProxy(options.ToKubeconfig)
	cTo, err := pTo.NewClient()
	if err != nil {
		return errors.Wrap(err, "failed to create controller-runtime client")
	}

	fmt.Print("hit enter to pause BMHs")
	in := bufio.NewScanner(os.Stdin)
	in.Scan()
	fmt.Println(in.Text())
	// Pause
	fromHosts, err := getBMHs(ctx, cFrom, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list BMH objects")
	}
	for _, fromHost := range fromHosts.Items {
		annotations := fromHost.GetAnnotations()
		if annotations == nil {
			fromHost.Annotations = map[string]string{}
		}
		fromHost.Annotations[bmh.PausedAnnotation] = "true"
		if err := cFrom.Update(ctx, &fromHost); err != nil {
			return errors.Wrapf(err, "error updating bmh %q %s/%s",
				fromHost.GroupVersionKind(), fromHost.GetNamespace(), fromHost.GetName())
		}
	}

	fmt.Print("hit enter to start clusterctl move (*make sure that you have added clusterctl labels to BMH CRD*)")
	in = bufio.NewScanner(os.Stdin)
	in.Scan()
	fmt.Println(in.Text())

	// Do clusterctl move
	clusterctlConfigPath := filepath.Join(util.GetRepositoryPath(config.ArtifactsPath), util.CLUSTERCTL_CONFIG_FILENAME)

	cctlClient, err := clusterctlclient.New(clusterctlConfigPath)
	if err != nil {
		return errors.Wrapf(err, "error creating clusterctl client")
	}
	if err := cctlClient.Move(clusterctlclient.MoveOptions{
		FromKubeconfig: options.FromKubeconfig,
		ToKubeconfig:   options.ToKubeconfig,
		Namespace:      options.Namespace,
	}); err != nil {
		return errors.Wrapf(err, "error during clusterctl move")
	}

	fmt.Print("hit enter update BMH Status")
	in = bufio.NewScanner(os.Stdin)
	in.Scan()
	fmt.Println(in.Text())
	hosts, err := getBMHs(ctx, cTo, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list BMH objects")
	}
	// Copy the Status field from old BMH to new BMH
	for _, host := range hosts.Items {
		t := metav1.Now()
		for _, fromHost := range fromHosts.Items {
			if fromHost.Name == host.Name {
				host.Status = fromHost.Status
				log.V(5).Info("Updating status on host %s/%s", host.Name, host.Namespace)
				break
			}
		}
		host.Status.LastUpdated = &t
		err = cTo.Status().Update(ctx, &host)
	}

	fmt.Print("hit enter to modify BMH and Node with new ProviderID")
	in = bufio.NewScanner(os.Stdin)
	in.Scan()
	fmt.Println(in.Text())
	// Check the ProviderID on the Metal3Machine, if it points to old BMH UID, then update to point to new BMH UID
	hosts, err = getBMHs(ctx, cTo, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list BMH objects")
	}
	for _, host := range hosts.Items {
		capm3Machine, err := getMetal3MachineByName(ctx, cTo, host.Spec.ConsumerRef.Name, host.Spec.ConsumerRef.Namespace)
		if err != nil {
			return errors.Wrapf(err, "failed to fetch Metal3Machine %s/%s for host %q %s/%s", host.Spec.ConsumerRef.Name, host.Spec.ConsumerRef.Namespace, host.GroupVersionKind(), host.GetNamespace(), host.GetName())
		}
		log.V(5).Info("UID of BMH %s and Provider ID on Meta3Machine %s ", host.UID, capm3Machine.Spec.ProviderID)
		newUIDURL := fmt.Sprintf("metal3://%s", host.UID)
		if *capm3Machine.Spec.ProviderID != newUIDURL {
			capm3Machine.Spec.ProviderID = &newUIDURL
			if err := updateMetal3Machine(ctx, cTo, capm3Machine); err != nil {
				return errors.Wrap(err, "failed to update Metal3Machine object")
			}
		} else {
			log.V(5).Info("UID on BMH matches that of the Metal3Machine, no update necessary")
		}
		// Update the Node
		nodeList := apicorev1.NodeList{}
		if err := cTo.List(ctx, &nodeList); err != nil {
			return errors.Wrap(err, "failed to list Node objects")
		}
		for _, node := range nodeList.Items {
			nodeProviderID, err := noderefutil.NewProviderID(node.Spec.ProviderID)
			if err != nil {
				return errors.Wrap(err, "Failed to parse ProviderID on node")
			}
			newUIDURL := fmt.Sprintf("metal3://%s", host.UID)
			if nodeProviderID.String() != newUIDURL {
				log.V(5).Info("updating provider-id on node %s from %s to %s", node.Name, nodeProviderID.ID, host.UID)
				node.Spec.ProviderID = newUIDURL
				if err := cTo.Update(ctx, &node); err != nil {
					return errors.Wrapf(err, "error updating Node %q %s/%s",
						node.GroupVersionKind(), node.GetNamespace(), node.GetName())
				}
			}
		}
	}

	// Remove Pause on new BMH
	fmt.Print("hit enter to remove pause on BMHs")
	in = bufio.NewScanner(os.Stdin)
	in.Scan()
	fmt.Println(in.Text())
	hosts, err = getBMHs(ctx, cTo, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list BMH objects")
	}
	for _, host := range hosts.Items {
		delete(host.Annotations, bmh.PausedAnnotation)
		if err := cTo.Update(ctx, &host); err != nil {
			return errors.Wrapf(err, "error updating bmh %q %s/%s",
				host.GroupVersionKind(), host.GetNamespace(), host.GetName())
		}
	}
	return nil
}

func getBMHs(ctx context.Context, c client.Client, namespace string) (bmh.BareMetalHostList, error) {
	hosts := bmh.BareMetalHostList{}
	opts := &client.ListOptions{
		Namespace: namespace,
	}

	err := c.List(ctx, &hosts, opts)
	return hosts, err
}

func getMetal3MachineByName(ctx context.Context, c client.Client, name, namespace string) (*capm3.Metal3Machine, error) {
	capm3Machine := &capm3.Metal3Machine{}
	objKey := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}
	if err := c.Get(ctx, objKey, capm3Machine); err != nil {
		return nil, err
	}
	return capm3Machine, nil
}

func updateMetal3Machine(ctx context.Context, c client.Client, capm3Machine *capm3.Metal3Machine) error {
	if err := c.Update(ctx, capm3Machine); err != nil {
		return errors.Wrapf(err, "error updating Metal3Machine %q %s/%s",
			capm3Machine.GroupVersionKind(), capm3Machine.GetNamespace(), capm3Machine.GetName())
	}
	return nil
}
