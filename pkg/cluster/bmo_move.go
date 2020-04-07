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
	"context"

	"github.com/Arvinderpal/metal3ctl/config"
	"github.com/Arvinderpal/metal3ctl/pkg/internal/proxy"
	bmh "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/cluster-api/cmd/clusterctl/log"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MoveBMOComponents(ctx context.Context, conf *config.Metal3CtlConfig, options *MoveOptions) error {
	log := logf.Log

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

	// Get the BMH objects from the 'from-cluster'
	hosts := bmh.BareMetalHostList{}
	opts := &client.ListOptions{
		Namespace: options.Namespace,
	}

	err = cFrom.List(ctx, &hosts, opts)
	if err != nil {
		return errors.Wrap(err, "failed to list BMH objects")
	}

	// Apply pause annotation on all BMH objects
	for _, host := range hosts.Items {
		log.V(5).Info("Object already exists, updating host %q %s/%s", host.GroupVersionKind(), host.GetNamespace(), host.GetName())
		annotations := host.GetAnnotations()
		if annotations == nil {
			host.Annotations = map[string]string{}
		}
		host.Annotations[bmh.PausedAnnotation] = "true"
		host.ResourceVersion = ""
		// Create BMH objects in the 'to-cluster'
		if err := cTo.Create(ctx, &host); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return errors.Wrap(err, "failed to create BMH objects in target cluster")
			} else {
				log.V(5).Info("Object already exists, updating to:", host)

			}
		}
		// At this point, we update the created host to exactly match bootstrap host
		// Retrieve the UID and the resource version for the update.
		existingHost := &bmh.BareMetalHost{}
		key := client.ObjectKey{
			Namespace: host.Namespace,
			Name:      host.Name,
		}
		if err := cTo.Get(ctx, key, existingHost); err != nil {
			return errors.Wrapf(err, "error reading resource for %q %s/%s",
				existingHost.GroupVersionKind(), existingHost.GetNamespace(), existingHost.GetName())
		}
		host.SetUID(existingHost.GetUID())
		host.SetResourceVersion(existingHost.GetResourceVersion())
		if err := cTo.Update(ctx, &host); err != nil {
			return errors.Wrapf(err, "error updating bmh %q %s/%s",
				host.GroupVersionKind(), host.GetNamespace(), host.GetName())
		}

		// Update BMH Status to match the `from-cluser`
		t := metav1.Now()
		host.Status.LastUpdated = &t
		err = cTo.Status().Update(ctx, &host)

		// Remove Pause
		// We first fetch the updated BMH
		// existingHost := &bmh.BareMetalHost{}
		// key := client.ObjectKey{
		// 	Namespace: host.Namespace,
		// 	Name:      host.Name,
		// }
		// if err := cTo.Get(ctx, key, existingHost); err != nil {
		// 	return errors.Wrapf(err, "error reading resource for %q %s/%s",
		// 		existingHost.GroupVersionKind(), existingHost.GetNamespace(), existingHost.GetName())
		// }
		// delete(host.Annotations, bmh.PausedAnnotation)
		// host.SetUID(existingHost.GetUID())
		// host.SetResourceVersion(existingHost.GetResourceVersion())
		// if err := cTo.Update(ctx, &host); err != nil {
		// 	return errors.Wrapf(err, "error updating bmh %q %s/%s",
		// 		host.GroupVersionKind(), host.GetNamespace(), host.GetName())
		// }
	}

	return nil
}
