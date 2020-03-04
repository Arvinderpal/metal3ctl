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
	"io/ioutil"

	"github.com/Arvinderpal/metal3ctl/config"
	"github.com/Arvinderpal/metal3ctl/pkg/internal/proxy"
	"github.com/Arvinderpal/metal3ctl/pkg/internal/util"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BMOConfig is the BMO config file that point to the repository created by InstallBMOComponents.
type BMOConfig struct {
	RawYAML   []byte
	Files     map[string][]byte
	Variables map[string]interface{}
}

func InstallBMOComponents(ctx context.Context, conf *config.Metal3CtlConfig) (*BMOConfig, error) {

	provider := conf.BMOProvider
	version := provider.Versions[0]
	// generate component yamls
	generator := config.ComponentGeneratorForComponentSource(version)
	manifest, err := generator.Manifests(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "error generating the manifest for %q / %q", provider.Name, version.Name)
	}

	fileMap := make(map[string][]byte)
	for _, file := range provider.Files {
		data, err := ioutil.ReadFile(file.SourcePath)
		if err != nil {
			return nil, errors.Wrapf(err, "error reading file %q / %q", provider.Name, file.SourcePath)
		}
		fileMap[file.TargetName] = data
	}

	// TODO: change to BMOVariables
	variablesMap := make(map[string]interface{})
	for key, value := range conf.Variables {
		variablesMap[key] = value
	}

	// transform the manifest to a list of objects
	objs, err := util.ToUnstructured(manifest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse yaml")
	}

	err = createComponents(ctx, proxy.NewProxy(conf.Kubeconfig), objs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bmo components in mgmt cluster")
	}

	return &BMOConfig{
		RawYAML:   manifest,
		Files:     fileMap,
		Variables: variablesMap,
	}, nil
}

func createComponents(ctx context.Context, p *proxy.Proxy, objs []unstructured.Unstructured) error {

	c, err := p.NewClient()
	if err != nil {
		return errors.Wrap(err, "failed to create controller-runtime client")
	}

	for i := range objs {
		obj := objs[i]

		// check if the component already exists, and eventually update it
		currentR := &unstructured.Unstructured{}
		currentR.SetGroupVersionKind(obj.GroupVersionKind())

		key := client.ObjectKey{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		}
		if err := c.Get(ctx, key, currentR); err != nil {
			if !apierrors.IsNotFound(err) {
				return errors.Wrapf(err, "failed to get current provider object")
			}

			//if it does not exists, create the component
			// log.V(5).Info("Creating", logf.UnstructuredToValues(obj)...)
			if err := c.Create(ctx, &obj); err != nil {
				return errors.Wrapf(err, "failed to create provider object %s, %s/%s", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
			}
			continue
		}

		// otherwise update the component
		// NB. we are using client.Merge PatchOption so the new objects gets compared with the current one server side
		// log.V(5).Info("Patching", logf.UnstructuredToValues(obj)...)
		obj.SetResourceVersion(currentR.GetResourceVersion())
		if err := c.Patch(ctx, &obj, client.Merge); err != nil {
			return errors.Wrapf(err, "failed to patch provider object")
		}
	}

	return nil
}

func DeleteBMOComponents(ctx context.Context, conf *config.Metal3CtlConfig, options *DeleteOptions) error {
	// TODO: Instead of gettitng the objects from the config file, we should instead get them from the cluster itself (see clusterctl approach).
	// TODO: support --include-crd and --include-namespaces

	version := conf.BMOProvider.Versions[0]
	// generate component yamls
	generator := config.ComponentGeneratorForComponentSource(version)
	manifest, err := generator.Manifests(ctx)
	if err != nil {
		return errors.Wrapf(err, "error generating the manifest for %q / %q", conf.BMOProvider.Name, version.Name)
	}
	// transform the manifest to a list of objects
	objs, err := util.ToUnstructured(manifest)
	if err != nil {
		return errors.Wrap(err, "failed to parse yaml")
	}
	err = deleteComponents(ctx, proxy.NewProxy(conf.Kubeconfig), objs)
	if err != nil {
		return errors.Wrap(err, "failed to create bmo components in mgmt cluster")
	}
	return nil
}

func deleteComponents(ctx context.Context, p *proxy.Proxy, objs []unstructured.Unstructured) error {

	c, err := p.NewClient()
	if err != nil {
		return errors.Wrap(err, "failed to create controller-runtime client")
	}
	errList := []error{}
	for _, obj := range objs {
		// log.V(5).Info("Deleting", logf.UnstructuredToValues(obj)...)
		if err := c.Delete(ctx, &obj); err != nil {
			if apierrors.IsNotFound(err) {
				// Tolerate IsNotFound error that might happen because we are not enforcing a deletion order
				continue
			}
			errList = append(errList, errors.Wrapf(err, "Error deleting object %s, %s/%s", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName()))
		}
	}
	return kerrors.NewAggregate(errList)
}
