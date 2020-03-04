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
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Arvinderpal/metal3ctl/config"
	metal3ctlinit "github.com/Arvinderpal/metal3ctl/pkg/init"
)

var io = &metal3ctlinit.InitOptions{}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a management cluster for metal-3",
	Long:  LongDesc(`TODO: add description`),

	Example: Examples(`
		# Initialize a management cluster using the metal3ctl configuration file.
		metal3ctl init

		# Lists the container images required for initializing the management cluster (without actually installing the providers).
		metal3ctl init  --list-images

		# Skips the baremetal-operator initialization.
		metal3ctl init  --skip-bmo

		# Skips the cluster-api component initialization.
		metal3ctl init  --skip-capi`),
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func init() {
	initCmd.Flags().BoolVarP(&io.ListImages, "list-images", "", false, "Lists the container images required for initializing the management cluster (without actually installing the providers)")
	initCmd.Flags().BoolVarP(&io.SkipBMO, "skip-bmo", "", false, "Skips the baremetal-operator initialization on the management cluster)")
	initCmd.Flags().BoolVarP(&io.SkipCAPI, "skip-capi", "", false, "Skips the cluster-api initialization on the management cluster)")
	RootCmd.AddCommand(initCmd)
}

func runInit() error {
	var err error

	metal3ctlCfgFile, err = filepath.Abs(metal3ctlCfgFile)
	if err != nil {
		return errors.Errorf("error converting %s to an absolute path", metal3ctlCfgFile)
	}

	configData, err := ioutil.ReadFile(metal3ctlCfgFile)
	if err != nil {
		return errors.Wrapf(err, "error reading the config file")
	}

	if io.ListImages {
		// TODO
		// images, err := cctlClient.InitImages(options)
		// if err != nil {
		// 	return err
		// }
		// for _, i := range images {
		// 	fmt.Println(i)
		// }
		// return nil
	}

	err = metal3ctlinit.InitMgmtCluster(config.LoadMetal3CtlConfigInput{ConfigData: configData}, io)
	if err != nil {
		return errors.Wrapf(err, "error while initializing management cluster")
	}

	return nil
}
