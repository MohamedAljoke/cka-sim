module github.com/MohamedAljoke/cka-sim/apps/platform

go 1.26.0

toolchain go1.26.9

require (
	github.com/MohamedAljoke/cka-sim/apps/go-shared v0.0.0
	sigs.k8s.io/yaml v1.4.0
)

replace github.com/MohamedAljoke/cka-sim/apps/go-shared => ../go-shared
