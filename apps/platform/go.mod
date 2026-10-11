module github.com/MohamedAljoke/cka-sim/apps/platform

go 1.26.0

toolchain go1.26.9

require (
	github.com/coder/websocket v1.8.15
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

require (
	github.com/MohamedAljoke/cka-sim/apps/go-shared v0.0.0
	github.com/MohamedAljoke/cka-sim/apps/lab v0.0.0
)

// The lab lives next door in this repo; drop this once it has its own.
replace github.com/MohamedAljoke/cka-sim/apps/lab => ../lab

replace github.com/MohamedAljoke/cka-sim/apps/go-shared => ../go-shared
