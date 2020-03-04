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

package proxy

import (
	"fmt"

	"github.com/Arvinderpal/metal3ctl/pkg/internal/scheme"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Scheme = scheme.Scheme
)

type Proxy struct {
	kubeconfig string
}

func (k *Proxy) NewClient() (client.Client, error) {
	config, err := k.getConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create controller-runtime client")
	}

	c, err := client.New(config, client.Options{Scheme: Scheme})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create controller-runtime client")
	}

	return c, nil
}

func NewProxy(kubeconfig string) *Proxy {
	// If a kubeconfig file isn't provided, find one in the standard locations.
	if kubeconfig == "" {
		kubeconfig = clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
	}
	return &Proxy{
		kubeconfig: kubeconfig,
	}
}

func (k *Proxy) getConfig() (*rest.Config, error) {
	config, err := clientcmd.LoadFromFile(k.kubeconfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load Kubeconfig file from %q", k.kubeconfig)
	}

	restConfig, err := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to rest client")
	}
	restConfig.UserAgent = fmt.Sprintf("metal3ctl")

	// Set QPS and Burst to a threshold that ensures the controller runtime client/client go does't generate throttling log messages
	restConfig.QPS = 20
	restConfig.Burst = 100

	return restConfig, nil
}
