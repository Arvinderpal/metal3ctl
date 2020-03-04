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

package delete

import (
	"context"
	"path/filepath"

	"github.com/Arvinderpal/metal3ctl/config"
	"github.com/Arvinderpal/metal3ctl/pkg/internal/util"
	"github.com/pkg/errors"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/pkg/client"
	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/pkg/client/config"
)

type DeleteOptions struct {
	IncludeNamespace bool
	IncludeCRDs      bool
}

func DeleteFromMgmtCluster(input config.LoadMetal3CtlConfigInput, dd *DeleteOptions) error {
	ctx := context.TODO()
	config, err := config.LoadMetal3CtlConfig(ctx, input)
	if err != nil {
		return errors.Wrapf(err, "error loading metal3ctl config file")
	}

	clusterctlConfigPath := filepath.Join(util.GetRepositoryPath(config.ArtifactsPath), util.CLUSTERCTL_CONFIG_FILENAME)

	cctlClient, err := clusterctlclient.New(clusterctlConfigPath)
	if err != nil {
		return errors.Wrapf(err, "error creating clusterctl client")
	}

	if err := cctlClient.Delete(clusterctlclient.DeleteOptions{
		Kubeconfig:              config.Kubeconfig,
		IncludeNamespace:        dd.IncludeNamespace,
		IncludeCRDs:             dd.IncludeCRDs,
		CoreProvider:            clusterctlconfig.ClusterAPIProviderName,
		BootstrapProviders:      []string{clusterctlconfig.KubeadmBootstrapProviderName},
		ControlPlaneProviders:   []string{clusterctlconfig.KubeadmControlPlaneProviderName},
		InfrastructureProviders: []string{config.InfraProvider()},
		DeleteAll:               true,
	}); err != nil {
		return errors.Wrapf(err, "error during clusterctl delete")
	}

	return nil
}
