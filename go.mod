module github.com/Arvinderpal/metal3ctl

go 1.13

require (
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/spf13/cobra v0.0.6
	sigs.k8s.io/cluster-api v0.0.0-00010101000000-000000000000
)

replace sigs.k8s.io/cluster-api => ../../../sigs.k8s.io/cluster-api
