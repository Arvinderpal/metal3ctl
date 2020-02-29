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

package config

import (
	"context"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/Arvinderpal/metal3ctl/config/exec"
	"github.com/pkg/errors"
)

// ContainerImage describes an image to load into a cluster and the behavior
// when loading the image.
type ContainerImage struct {
	// Name is the fully qualified name of the image.
	Name string
}

// ComponentSourceType indicates how a component's source should be obtained.
type ComponentSourceType string

const (
	// URLSource is component YAML available directly via a URL.
	// The URL may begin with file://, http://, or https://.
	URLSource ComponentSourceType = "url"

	// KustomizeSource is a valid kustomization root that can be used to produce
	// the component YAML.
	KustomizeSource ComponentSourceType = "kustomize"
)

// ComponentSource describes how to obtain a component's YAML.
type ComponentSource struct {
	// Name is used for logging when a component has multiple sources.
	Name string `json:"name,omitempty"`

	// Value is the source of the component's YAML.
	// May be a URL or a kustomization root (specified by Type).
	// If a Type=url then Value may begin with file://, http://, or https://.
	// If a Type=kustomize then Value may be any valid go-getter URL. For
	// more information please see https://github.com/hashicorp/go-getter#url-format.
	Value string `json:"value"`

	// Type describes how to process the source of the component's YAML.
	//
	// Defaults to "kustomize".
	Type ComponentSourceType `json:"type,omitempty"`

	// Replacements is a list of patterns to replace in the component YAML
	// prior to application.
	Replacements []ComponentReplacement `json:"replacements,omitempty"`
}

// ComponentWaiterType indicates the type of check to use to determine if the
// installed components are ready.
type ComponentWaiterType string

const (
	// ServiceWaiter indicates to wait until a service's condition is Available.
	// When ComponentWaiter.Value is set to "service", the ComponentWaiter.Value
	// should be set to the name of a Service resource.
	ServiceWaiter ComponentWaiterType = "service"

	// PodsWaiter indicates to wait until all the pods in a namespace have a
	// condition of Ready.
	// When ComponentWaiter.Value is set to "pods", the ComponentWaiter.Value
	// should be set to the name of a Namespace resource.
	PodsWaiter ComponentWaiterType = "pods"
)

// ComponentWaiter contains information to help determine whether installed
// components are ready.
type ComponentWaiter struct {
	// Value varies depending on the specified Type.
	// Please see the documentation for the different WaiterType constants to
	// understand the valid values for this field.
	Value string `json:"value"`

	// Type describes the type of check to perform.
	//
	// Defaults to "pods".
	Type ComponentWaiterType `json:"type,omitempty"`
}

// ComponentReplacement is used to replace some of the generated YAML prior
// to application.
type ComponentReplacement struct {
	// Old is the pattern to replace.
	// A regular expression may be used.
	Old string `json:"old"`
	// New is the string used to replace the old pattern.
	// An empty string is valid.
	New string `json:"new,omitempty"`
}

// ComponentConfig describes a required component.
type ComponentConfig struct {
	// Name is the name of the component.
	// This field is primarily used for logging.
	Name string `json:"name"`

	// Sources is an optional list of component YAML to apply to the management
	// cluster.
	// This field may be omitted when wanting only to block progress via one or
	// more Waiters.
	Sources []ComponentSource `json:"sources,omitempty"`

	// Waiters is an optional list of checks to perform in order to determine
	// whether or not the installed components are ready.
	Waiters []ComponentWaiter `json:"waiters,omitempty"`
}

// ProviderConfig describes a provider to be configured in the local repository that will be created.
type ProviderConfig struct {
	// Name is the name of the provider.
	Name string `json:"name"`

	// Type is the type of the provider.
	Type string `json:"type"`

	// Versions is a list of component YAML to be added to the local repository, one for each release.
	// Please note that the first source will be used as a default release for this provider.
	Versions []ComponentSource `json:"versions,omitempty"`

	// Files is a list of files to be copied into the local repository for the default release of this provider.
	Files []Files `json:"files,omitempty"`

	// Waiters is list of waiters to be used to check if the installed provider are ready.
	Waiters []ProviderWaiter `json:"waiters,omitempty"`
}

// ProviderWaiterType indicates the type of check to use to determine if the
// installed provider are ready.
type ProviderWaiterType string

const (
	// ApiServiceWaiter indicates to wait until an apiservice have a
	// condition of Ready.
	// When ComponentWaiter.Value is set to "apiservice", the ComponentWaiter.Name
	// should be set to the name of a api resource resource.
	ApiServiceWaiter ProviderWaiterType = "apiservice"

	// DeploymentWaiter indicates to wait until a deployment have a
	// condition of Ready.
	// When ComponentWaiter.Value is set to "deployment", the ComponentWaiter.Name
	// should be set to the name of a deployment resource.
	// You should use one of ComponentWaiter.Namespace or ComponentWaiter.DefaultNamespace to specify the target
	// namespace of the deployment; if the target namespace passed to clusterctl init is not empty, it will override
	// ComponentWaiter.DefaultNamespace.
	DeploymentWaiter ProviderWaiterType = "deployment"
)

// ProviderWaiter contains information to help determine whether installed
// provider are ready.
type ProviderWaiter struct {
	// Type describes the type of check to perform.
	// Defaults to "deployment".
	Type ProviderWaiterType `json:"type,omitempty"`

	// Namespace varies depending on the specified Type.
	// Please see the documentation for the different WaiterType constants to
	// understand the valid values for this field.
	Namespace string `json:"namespace,omitempty"`

	// DefaultNamespace varies depending on the specified Type.
	// Please see the documentation for the different WaiterType constants to
	// understand the valid values for this field.
	DefaultNamespace string `json:"defaultNamespace,omitempty"`

	// Name varies depending on the specified Type.
	// Please see the documentation for the different WaiterType constants to
	// understand the valid values for this field.
	Name string `json:"name"`
}

// Files contains information about files to be copied into the local repository
type Files struct {
	// SourcePath path of the file.
	SourcePath string `json:"sourcePath"`

	// TargetName name of the file copied into the local repository. if empty, the source name
	// Will be preserved
	TargetName string `json:"targetName,omitempty"`
}

// YAMLForComponentSource returns the YAML for the provided component source.
func YAMLForComponentSource(ctx context.Context, source ComponentSource) ([]byte, error) {
	var data []byte

	switch source.Type {
	case URLSource:
		resp, err := http.Get(source.Value)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		data = buf
	case KustomizeSource:
		kustomize := exec.NewCommand(
			exec.WithCommand("kustomize"),
			exec.WithArgs("build", source.Value))
		stdout, stderr, err := kustomize.Run(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to execute kustomize: %s", stderr)
		}
		data = stdout
	default:
		return nil, errors.Errorf("invalid type: %q", source.Type)
	}

	for _, replacement := range source.Replacements {
		rx, err := regexp.Compile(replacement.Old)
		if err != nil {
			return nil, err
		}
		data = rx.ReplaceAll(data, []byte(replacement.New))
	}

	return data, nil
}

// ComponentGenerator is used to install components, generally any YAML bundle.
type ComponentGenerator interface {
	// GetName returns the name of the component.
	GetName() string
	// Manifests return the YAML bundle.
	Manifests(context.Context) ([]byte, error)
}

// ComponentGeneratorForComponentSource returns a ComponentGenerator for the
// provided ComponentSource.
func ComponentGeneratorForComponentSource(source ComponentSource) ComponentGenerator {
	return componentSourceGenerator{ComponentSource: source}
}

type componentSourceGenerator struct {
	ComponentSource
}

// GetName returns the name of the component.
func (g componentSourceGenerator) GetName() string {
	return g.Name
}

// Manifests return the YAML bundle.
func (g componentSourceGenerator) Manifests(ctx context.Context) ([]byte, error) {
	return YAMLForComponentSource(ctx, g.ComponentSource)
}
