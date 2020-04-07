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

package scheme

import (
	bmoapis "github.com/metal3-io/baremetal-operator/pkg/apis"
	infrav1alpha2 "github.com/metal3-io/cluster-api-provider-metal3/api/v1alpha2"
	infrav1alpha3 "github.com/metal3-io/cluster-api-provider-metal3/api/v1alpha3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
)

var (
	// Scheme contains a set of API resources used by clusterctl
	Scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(Scheme)
	_ = clusterctlv1.AddToScheme(Scheme)
	_ = clusterv1.AddToScheme(Scheme)
	_ = apiextensionsv1.AddToScheme(Scheme)
	_ = bmoapis.AddToScheme(Scheme)
	_ = infrav1alpha2.AddToScheme(Scheme)
	_ = infrav1alpha3.AddToScheme(Scheme)
}
