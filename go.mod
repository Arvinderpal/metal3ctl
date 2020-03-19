module github.com/Arvinderpal/metal3ctl

go 1.13

require (
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.6
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/cluster-api v0.2.6-0.20200213153035-a0cdb3b05cda
	sigs.k8s.io/cluster-api/test/framework v0.0.0-20200226161228-25b281e30f43 // indirect
	sigs.k8s.io/controller-runtime v0.5.1
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/cluster-api => ../../../sigs.k8s.io/cluster-api
