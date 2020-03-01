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

	"github.com/Arvinderpal/metal3ctl/config"
	"github.com/pkg/errors"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/pkg/client"
	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/pkg/client/config"
)

func InitMgmtCluster(input config.LoadMetal3CtlConfigInput) error {
	ctx := context.TODO()
	config, err := config.LoadMetal3CtlConfig(ctx, input)
	if err != nil {
		return errors.Wrapf(err, " error loading metal3ctl config file")
	}
	// prettyJSON, err := json.MarshalIndent(config, "", "    ")
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("%s\n", string(prettyJSON))

	// Creates a local provider repository based on the configuration and a clusterctl config file that reads from this repository.
	clusterctlConfig, err := CreateRepository(ctx, CreateRepositoryInput{
		config:        config,
		artifactsPath: config.ArtifactPath,
	})
	if err != nil {
		return err
	}

	cctlClient, err := clusterctlclient.New(clusterctlConfig.Path)
	if err != nil {
		return err
	}

	initOpt := clusterctlclient.InitOptions{
		Kubeconfig:              config.Kubeconfig,
		CoreProvider:            clusterctlconfig.ClusterAPIProviderName,
		BootstrapProviders:      []string{clusterctlconfig.KubeadmBootstrapProviderName},
		ControlPlaneProviders:   []string{clusterctlconfig.KubeadmControlPlaneProviderName},
		InfrastructureProviders: []string{config.InfraProvider()},
	}

	_, err = cctlClient.Init(initOpt)
	if err != nil {
		return errors.Wrap(err, "failed to run clusterctl init")
	}

	return nil
}
