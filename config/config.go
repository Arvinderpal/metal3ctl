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
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/pkg/client/config"
	"sigs.k8s.io/yaml"
)

// LoadMetal3CtlConfig is the input for LoadMetal3CtlConfig.
type LoadMetal3CtlConfigInput struct {
	ConfigData []byte
}

// LoadMetal3CtlConfig will load the metal3ctl config.
func LoadMetal3CtlConfig(ctx context.Context, input LoadMetal3CtlConfigInput) (*Metal3CtlConfig, error) {
	if len(input.ConfigData) == 0 {
		return nil, errors.New("config should not be empty")
	}

	config := &Metal3CtlConfig{}
	if err := yaml.Unmarshal(input.ConfigData, config); err != nil {
		return nil, errors.Wrapf(err, "error loading the init config file")
	}

	config.Defaults()
	if err := config.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid init config")
	}
	return config, nil
}

// Metal3CtlConfig is the input used to configure a metal3 mgmt cluster.
type Metal3CtlConfig struct {
	// Name is the name of the management cluster.
	ManagementClusterName string `json:"managementClusterName,omitempty"`

	// Path to kubeconfig of the mgmt cluster.
	Kubeconfig string `json:"kubeconfig,omitempty"`

	// Path to where all the generated artifacts will be stored.
	ArtifactsPath string `json:"artifactsPath,omitempty"`

	// Images is a list of container images to load into the mgmt cluster.
	Images []ContainerImage `json:"images,omitempty"`

	// CAPIProviders is a list of cluster-api providers to be configured in the local repository that will then be created on the mgmt cluster.
	// It is required to provide following providers
	// - cluster-api
	// - bootstrap kubeadm
	// - control-plane kubeadm
	// - cluster-api-baremetal
	CAPIProviders []ProviderConfig `json:"capiProviders,omitempty"`

	// baremetal-operator configuration
	BMOProvider ProviderConfig `json:"bmoProvider,omitempty"`

	// Variables to be added to the clusterctl config file
	// Please not that clusterctl read variables from the os environment variables as well, so you can avoid to hard code
	// sensitive data in the config file.
	Variables map[string]string `json:"variables,omitempty"`
}

// Defaults assigns default values to the object.
func (c *Metal3CtlConfig) Defaults() {

	for i := range c.CAPIProviders {
		provider := &c.CAPIProviders[i]
		for j := range provider.Versions {
			version := &provider.Versions[j]
			if version.Value != "" && version.Type == "" {
				version.Type = KustomizeSource
			}
		}
		for j := range provider.Files {
			file := &provider.Files[j]
			if file.SourcePath != "" && file.TargetName == "" {
				file.TargetName = filepath.Base(file.SourcePath)
			}
		}
		for j := range provider.Waiters {
			waiter := &provider.Waiters[j]
			if waiter.Type == "" {
				waiter.Type = DeploymentWaiter
			}
		}
	}
}

// Validate validates the configuration.
func (c *Metal3CtlConfig) Validate() error {
	if c.ManagementClusterName == "" {
		return errEmptyArg("managementClusterName")
	}
	if c.Kubeconfig == "" {
		return errEmptyArg("kubeconfig")
	}
	if c.ArtifactsPath == "" {
		return errEmptyArg("artifactsPath")
	}
	providersByType := map[clusterctlv1.ProviderType][]string{
		clusterctlv1.CoreProviderType:           nil,
		clusterctlv1.BootstrapProviderType:      nil,
		clusterctlv1.ControlPlaneProviderType:   nil,
		clusterctlv1.InfrastructureProviderType: nil,
	}
	for i, providerConfig := range c.CAPIProviders {
		if providerConfig.Name == "" {
			return errEmptyArg(fmt.Sprintf("CAPIProviders[%d].Name", i))
		}
		providerType := clusterctlv1.ProviderType(providerConfig.Type)
		switch providerType {
		case clusterctlv1.CoreProviderType, clusterctlv1.BootstrapProviderType, clusterctlv1.ControlPlaneProviderType, clusterctlv1.InfrastructureProviderType:
			providersByType[providerType] = append(providersByType[providerType], providerConfig.Name)
		default:
			return errInvalidArg("CAPIProviders[%d].Type=%q", i, providerConfig.Type)
		}

		for j, version := range providerConfig.Versions {
			if version.Name == "" {
				return errEmptyArg(fmt.Sprintf("CAPIProviders[%d].Sources[%d].Name", i, j))
			}
			switch version.Type {
			case URLSource, KustomizeSource:
				if version.Value == "" {
					return errEmptyArg(fmt.Sprintf("CAPIProviders[%d].Sources[%d].Value", i, j))
				}
			default:
				return errInvalidArg("CAPIProviders[%d].Sources[%d].Type=%q", i, j, version.Type)
			}
			for k, replacement := range version.Replacements {
				if _, err := regexp.Compile(replacement.Old); err != nil {
					return errInvalidArg("CAPIProviders[%d].Sources[%d].Replacements[%d].Old=%q: %v", i, j, k, replacement.Old, err)
				}
			}
		}

		for j, file := range providerConfig.Files {
			if file.SourcePath == "" {
				return errInvalidArg("CAPIProviders[%d].Files[%d].SourcePath=%q", i, j, file.SourcePath)
			}
			if !fileExists(file.SourcePath) {
				return errInvalidArg("CAPIProviders[%d].Files[%d].SourcePath=%q", i, j, file.SourcePath)
			}
			if file.TargetName == "" {
				return errInvalidArg("CAPIProviders[%d].Files[%d].TargetName=%q", i, j, file.TargetName)
			}
		}

		for j, waiter := range providerConfig.Waiters {
			switch waiter.Type {
			case ApiServiceWaiter:
				//TODO: add validation
			case DeploymentWaiter:
				//TODO: add validation
			default:
				return errInvalidArg("CAPIProviders[%d].Waiters[%d].Type=%q", i, j, waiter.Type)
			}
		}
	}

	if len(providersByType[clusterctlv1.CoreProviderType]) != 1 {
		return errInvalidArg("invalid config: it is required to have exactly one core-provider")
	}
	if providersByType[clusterctlv1.CoreProviderType][0] != clusterctlconfig.ClusterAPIProviderName {
		return errInvalidArg("invalid config: core-provider should be named %s", clusterctlconfig.ClusterAPIProviderName)
	}

	if len(providersByType[clusterctlv1.BootstrapProviderType]) != 1 {
		return errInvalidArg("invalid config: it is required to have exactly one bootstrap-provider")
	}
	if providersByType[clusterctlv1.BootstrapProviderType][0] != clusterctlconfig.KubeadmBootstrapProviderName {
		return errInvalidArg("invalid config: bootstrap-provider should be named %s", clusterctlconfig.KubeadmBootstrapProviderName)
	}

	if len(providersByType[clusterctlv1.ControlPlaneProviderType]) != 1 {
		return errInvalidArg("invalid config: it is required to have exactly one control-plane-provider")
	}
	if providersByType[clusterctlv1.ControlPlaneProviderType][0] != clusterctlconfig.KubeadmControlPlaneProviderName {
		return errInvalidArg("invalid config: control-plane-provider should be named %s", clusterctlconfig.KubeadmControlPlaneProviderName)
	}

	if len(providersByType[clusterctlv1.InfrastructureProviderType]) != 1 {
		return errInvalidArg("invalid config: it is required to have exactly one infrastructure-provider")
	}

	//TODO: check if the infrastructure provider has a cluster-template

	for i, containerImage := range c.Images {
		if containerImage.Name == "" {
			return errEmptyArg(fmt.Sprintf("Images[%d].Name=%q", i, containerImage.Name))
		}
	}

	// TODO: validate BMOProvider

	return nil
}

func errInvalidArg(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return errors.Errorf("invalid argument: %s", msg)
}

func errEmptyArg(argName string) error {
	return errInvalidArg("%s is empty", argName)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (c *Metal3CtlConfig) InfraProvider() string {
	for _, providerConfig := range c.CAPIProviders {
		if providerConfig.Type == string(clusterctlv1.InfrastructureProviderType) {
			return providerConfig.Name
		}
	}
	panic("it is required to have an infra provider in the config")
}
