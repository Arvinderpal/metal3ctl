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
	"io/ioutil"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// ClusterctlConfig is the clusterctl config file that point to the repository created by CreateRepository.
type ClusterctlConfig struct {
	Path   string
	Values map[string]interface{}
}

// ClusterctlConfigProvider mirrors clusterctl config.Provider interface and allows serialization of the corresponding info
type ClusterctlConfigProvider struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
	Type string `json:"type,omitempty"`
}

// WriteFile writes a clusterctl config file to disk.
func (c *ClusterctlConfig) WriteFile() error {
	data, err := yaml.Marshal(c.Values)
	if err != nil {
		return errors.Wrapf(err, "failed to convert to yaml the clusterctl config file")
	}
	if err := ioutil.WriteFile(c.Path, data, 0755); err != nil {
		return err
	}
	return nil
}
