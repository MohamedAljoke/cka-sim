module github.com/MohamedAljoke/cka-sim/apps/platform

go 1.26.0

toolchain go1.26.9

require (
	connectrpc.com/connect/v2 v2.0.1
	github.com/MohamedAljoke/cka-sim/apps/go-shared v0.0.0
	google.golang.org/protobuf v1.36.12
	sigs.k8s.io/yaml v1.4.0
)

replace github.com/MohamedAljoke/cka-sim/apps/go-shared => ../go-shared
