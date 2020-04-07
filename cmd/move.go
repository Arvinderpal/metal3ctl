/*
Copyright 2019 The Kubernetes Authors.

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

	"github.com/Arvinderpal/metal3ctl/config"
	metal3ctl "github.com/Arvinderpal/metal3ctl/pkg/cluster"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var mo = &metal3ctl.MoveOptions{}

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move BMH objects, Cluster API objects and all dependencies between management clusters.",
	Long: LongDesc(`
		Move BMH objects, Cluster API objects and all dependencies between management clusters.

		Note: The destination cluster MUST have the required provider components installed.`),

	Example: Examples(`
		Move BMH objects, Cluster API objects and all dependencies between management clusters.
		metal3ctl move --to-kubeconfig=target-kubeconfig.yaml

		# Skips the BMH move.
		metal3ctl move  --skip-bmo

		# Skips the clusterctl move.
		metal3ctl move  --skip-capi`),
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMove()
	},
}

func init() {
	moveCmd.Flags().StringVar(&mo.FromKubeconfig, "kubeconfig", "",
		"Path to the kubeconfig file for the source management cluster. If unspecified, default discovery rules apply.")
	moveCmd.Flags().StringVar(&mo.ToKubeconfig, "to-kubeconfig", "",
		"Path to the kubeconfig file to use for the destination management cluster.")
	moveCmd.Flags().StringVarP(&mo.Namespace, "namespace", "n", "",
		"The namespace where the workload cluster is hosted. If unspecified, the current context's namespace is used.")
	moveCmd.Flags().BoolVarP(&dd.SkipBMO, "skip-bmo", "", false, "Skips the move of BMH objects)")
	moveCmd.Flags().BoolVarP(&dd.SkipCAPI, "skip-capi", "", false, "Skips the move of cluster-api objects)")
	RootCmd.AddCommand(moveCmd)
}

func runMove() error {
	var err error
	if mo.ToKubeconfig == "" {
		return errors.New("please specify a target cluster using the --to-kubeconfig flag")
	}

	metal3ctlCfgFile, err = filepath.Abs(metal3ctlCfgFile)
	if err != nil {
		return errors.Errorf("error converting %s to an absolute path", metal3ctlCfgFile)
	}

	configData, err := ioutil.ReadFile(metal3ctlCfgFile)
	if err != nil {
		return errors.Wrapf(err, "error reading the config file")
	}

	err = metal3ctl.MoveFromBootstrapToTargetCluster(config.LoadMetal3CtlConfigInput{ConfigData: configData}, mo)
	if err != nil {
		return errors.Wrapf(err, "error while moving")
	}
	return nil
}
