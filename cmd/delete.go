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
	metal3ctldelete "github.com/Arvinderpal/metal3ctl/pkg/cluster"
)

var dd = &metal3ctldelete.DeleteOptions{}

var deleteCmd = &cobra.Command{
	Use:   "delete ",
	Short: "Deletes one or more providers from the management cluster",
	Long: LongDesc(`
		Deletes one or more providers from the management cluster.`),

	Example: Examples(`
		# Deletes the providers that were initialized with metal3ctl init
		# Please note that this implies the deletion of all provider components except the hosting namespace
		# and the CRDs.
		metal3ctl delete 

		# Reset the management cluster to its original state
		# Important! As a consequence of this operation all the corresponding resources on target clouds
		# are "orphaned" and thus there may be ongoing costs incurred as a result of this.
		metal3ctl delete --include-crd  --include-namespace
		
		# Skips the baremetal-operator deletion.
		metal3ctl init  --skip-bmo

		# Skips the cluster-api component deletion.
		metal3ctl init  --skip-capi`),
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDelete()
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&dd.IncludeNamespace, "include-namespace", "", false, "Forces the deletion of the namespace where the providers are hosted (and of all the contained objects)")
	deleteCmd.Flags().BoolVarP(&dd.IncludeCRDs, "include-crd", "", false, "Forces the deletion of the provider's CRDs (and of all the related objects)")
	initCmd.Flags().BoolVarP(&dd.SkipBMO, "skip-bmo", "", false, "Skips the baremetal-operator deletion on the management cluster)")
	initCmd.Flags().BoolVarP(&dd.SkipCAPI, "skip-capi", "", false, "Skips the cluster-api deletion on the management cluster)")
	RootCmd.AddCommand(deleteCmd)
}

func runDelete() error {
	var err error

	metal3ctlCfgFile, err = filepath.Abs(metal3ctlCfgFile)
	if err != nil {
		return errors.Errorf("error converting %s to an absolute path", metal3ctlCfgFile)
	}

	configData, err := ioutil.ReadFile(metal3ctlCfgFile)
	if err != nil {
		return errors.Wrapf(err, "error reading the config file")
	}

	err = metal3ctldelete.DeleteFromMgmtCluster(config.LoadMetal3CtlConfigInput{ConfigData: configData}, dd)
	if err != nil {
		return errors.Wrapf(err, "error while initializing management cluster")
	}

	return nil

}
