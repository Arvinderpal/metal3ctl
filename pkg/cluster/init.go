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
	"github.com/pkg/errors"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
)

type InitOptions struct {
	ListImages bool
	SkipBMO    bool
	SkipCAPI   bool
}

func InitMgmtCluster(input config.LoadMetal3CtlConfigInput, options *InitOptions) error {
	var err error
	ctx := context.TODO()
	config, err := config.LoadMetal3CtlConfig(ctx, input)
	if err != nil {
		return errors.Wrapf(err, " error loading metal3ctl config file")
	}

	// TODO: Prefetch Images into mgmt cluster.
	// TODO: This is minikube specific. Make it more generic.
	// sudo minikube ssh sudo docker pull quay.io/metal3-io/ironic
	// for _, containerImage := range config.Images {
	// 	fmt.Printf("Fetching image %q", containerImage.Name)
	// }
	if !options.SkipBMO {
		_, err = InstallBMOComponents(ctx, config)
		if err != nil {
			return errors.Wrapf(err, "error installing baremetal-operator components")
		}
	}

	if !options.SkipCAPI {
		// Creates a local provider repository based on the configuration and a clusterctl config file that reads from this repository.
		clusterctlConfig, err := CreateCAPIRepository(ctx, CreateCAPIRepositoryInput{
			config:        config,
			artifactsPath: config.ArtifactsPath,
		})
		if err != nil {
			return errors.Wrapf(err, "error creating local cluster-api repository")
		}

		cctlClient, err := clusterctlclient.New(clusterctlConfig.Path)
		if err != nil {
			return errors.Wrapf(err, "error creating clusterctl client")
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
	}

	return nil
}
