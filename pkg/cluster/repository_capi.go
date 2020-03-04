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
	"os"
	"path/filepath"

	"github.com/Arvinderpal/metal3ctl/config"
	"github.com/Arvinderpal/metal3ctl/pkg/internal/util"
	"github.com/pkg/errors"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
)

// CreateCAPIRepositoryInput is the input for CreateCAPIRepository.
type CreateCAPIRepositoryInput struct {
	artifactsPath string
	config        *config.Metal3CtlConfig
}

// CreateCAPIRepository creates a local repository based on the metal3ctl config and returns a metal3ctl config
// file to be used for working with such repository.
func CreateCAPIRepository(ctx context.Context, input CreateCAPIRepositoryInput) (*ClusterctlConfig, error) {
	providers := []ClusterctlConfigProvider{}
	repositoryPath := util.GetRepositoryPath(input.artifactsPath)
	for _, provider := range input.config.CAPIProviders {
		providerUrl := ""
		for _, version := range provider.Versions {
			providerLabel := clusterctlv1.ManifestLabel(provider.Name, clusterctlv1.ProviderType(provider.Type))

			generator := config.ComponentGeneratorForComponentSource(version)
			manifest, err := generator.Manifests(ctx)
			if err != nil {
				return nil, errors.Wrapf(err, "error generating the manifest for %q / %q", providerLabel, version.Name)
			}
			sourcePath := filepath.Join(repositoryPath, providerLabel, version.Name)
			if err := os.MkdirAll(sourcePath, 0755); err != nil {
				return nil, errors.Wrapf(err, "error creating the repository folder for %q / %q", providerLabel, version.Name)
			}

			filePath := filepath.Join(sourcePath, "components.yaml")
			if err := ioutil.WriteFile(filePath, manifest, 0755); err != nil {
				return nil, errors.Wrapf(err, "error writing manifest for %q / %q", providerLabel, version.Name)
			}

			if providerUrl == "" {
				providerUrl = filePath
			}
		}
		providers = append(providers, ClusterctlConfigProvider{
			Name: provider.Name,
			URL:  providerUrl,
			Type: provider.Type,
		})

		for _, file := range provider.Files {
			data, err := ioutil.ReadFile(file.SourcePath)
			if err != nil {
				return nil, errors.Wrapf(err, "error reading file %q / %q", provider.Name, file.SourcePath)
			}

			destinationFile := filepath.Join(filepath.Dir(providerUrl), file.TargetName)
			err = ioutil.WriteFile(destinationFile, data, 0644)
			if err != nil {
				return nil, errors.Wrapf(err, "error writing file %q / %q", provider.Name, file.TargetName)
			}
		}
	}

	clusterctlConfigFile := &ClusterctlConfig{
		Path: filepath.Join(repositoryPath, util.CLUSTERCTL_CONFIG_FILENAME),
		Values: map[string]interface{}{
			"providers": providers,
		},
	}
	for key, value := range input.config.Variables {
		clusterctlConfigFile.Values[key] = value
	}
	if err := clusterctlConfigFile.WriteFile(); err != nil {
		return nil, errors.Wrapf(err, "error writing the clusterctl config file")
	}
	return clusterctlConfigFile, nil
}
