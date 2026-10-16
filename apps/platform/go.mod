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
	github.com/golang-jwt/jwt/v5 v5.3.1
)

replace github.com/MohamedAljoke/cka-sim/apps/go-shared => ../go-shared
