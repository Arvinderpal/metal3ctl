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

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Arvinderpal/metal3ctl/config"
)

type initOptions struct {
	listImages bool
}

var io = &initOptions{}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a management cluster for metal-3",
	Long:  LongDesc(`TODO: add description`),

	Example: Examples(`
		# Initialize a management cluster using the metal3ctl configuration file.
		metal3ctl init --config=my-metal3ctl.config.yaml

		# Lists the container images required for initializing the management cluster (without actually installing the providers).
		metal3ctl init --config=my-metal3ctl.config.yaml --list-images`),
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func init() {
	initCmd.Flags().BoolVarP(&io.listImages, "list-images", "", false, "Lists the container images required for initializing the management cluster (without actually installing the providers)")

	RootCmd.AddCommand(initCmd)
}

func runInit() error {
	var err error

	metal3ctlCfgFile, err = filepath.Abs(metal3ctlCfgFile)
	if err != nil {
		return fmt.Errorf("error converting %s to an absolute path", metal3ctlCfgFile)
	}

	config, err := config.LoadMetal3CtlConfig(context.TODO(), config.LoadMetal3CtlConfigInput{
		ConfigPath: metal3ctlCfgFile,
	})
	if err != nil {
		return err
	}

	prettyJSON, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(prettyJSON))

	// cctlClient, err := client.New(clusterctlCfgFile)
	// if err != nil {
	// 	return err
	// }

	// options := client.InitOptions{
	// 	Kubeconfig:              io.kubeconfig,
	// 	CoreProvider:            io.coreProvider,
	// 	BootstrapProviders:      io.bootstrapProviders,
	// 	ControlPlaneProviders:   io.controlPlaneProviders,
	// 	InfrastructureProviders: io.infrastructureProviders,
	// 	TargetNamespace:         io.targetNamespace,
	// 	WatchingNamespace:       io.watchingNamespace,
	// 	LogUsageInstructions:    true,
	// }

	// if io.listImages {
	// 	images, err := cctlClient.InitImages(options)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	for _, i := range images {
	// 		fmt.Println(i)
	// 	}
	// 	return nil
	// }

	// if _, err := cctlClient.Init(options); err != nil {
	// 	return err
	// }
	return nil
}
