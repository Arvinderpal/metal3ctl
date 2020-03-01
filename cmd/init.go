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
		return errors.Errorf("error converting %s to an absolute path", metal3ctlCfgFile)
	}

	configData, err := ioutil.ReadFile(metal3ctlCfgFile)
	if err != nil {
		return errors.Wrapf(err, "error reading the config file")
	}

	if io.listImages {
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

	err = metal3ctlinit.InitMgmtCluster(config.LoadMetal3CtlConfigInput{ConfigData: configData})
	if err != nil {
		return errors.Wrapf(err, "error while initializing management cluster")
	}

	return nil
}
