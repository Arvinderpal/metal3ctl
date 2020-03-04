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

package init

import (
	"context"
	"io/ioutil"

	"github.com/Arvinderpal/metal3ctl/config"
	"github.com/Arvinderpal/metal3ctl/pkg/internal/proxy"
	"github.com/Arvinderpal/metal3ctl/pkg/internal/util"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InstallBMOComponentsInput is the input for installBMOComponents.
type InstallBMOComponentsInput struct {
	config *config.Metal3CtlConfig
}

// BMOConfig is the BMO config file that point to the repository created by InstallBMOComponents.
type BMOConfig struct {
	RawYAML   []byte
	Files     map[string][]byte
	Variables map[string]interface{}
}

func InstallBMOComponents(ctx context.Context, input InstallBMOComponentsInput) (*BMOConfig, error) {

	err := input.Validation()
	if err != nil {
		return nil, errors.Wrapf(err, "validation failed for baremetal-operator configuration")
	}
	provider := input.config.BMOProvider
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
	for key, value := range input.config.Variables {
		variablesMap[key] = value
	}

	// transform the manifest to a list of objects
	objs, err := util.ToUnstructured(manifest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse yaml")
	}

	err = createComponents(ctx, proxy.NewProxy(input.config.Kubeconfig), objs)
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

func DeleteBMOComponents() error {
	// TODO
	return nil
}

func (input InstallBMOComponentsInput) Validation() error {
	provider := input.config.BMOProvider
	if len(provider.Versions) != 1 {
		// TODO: consider adding support for multiple bmo versions
		return errors.New("please specify one and only one baremetal-operator version")
	}
	if provider.Name == "" {
		return errors.New("baremetal-operator name cannot be empty in metal3ctl configuration file")
	}
	if provider.Type != "BareMetalOperator" {
		return errors.Errorf("baremetal-operator type must be BareMetalOperator, found %v instead", provider.Type)
	}
	return nil
}
