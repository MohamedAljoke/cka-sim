module github.com/MohamedAljoke/cka-sim/apps/lab

go 1.26.0

toolchain go1.26.9

require (
	connectrpc.com/connect/v2 v2.0.1
	github.com/MohamedAljoke/cka-sim/apps/go-shared v0.0.0
	google.golang.org/protobuf v1.36.12
	sigs.k8s.io/kind v0.33.0
	sigs.k8s.io/yaml v1.4.0
)

require (
	al.essio.dev/pkg/shellescape v1.5.1 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.8.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.6.0 // indirect
)

replace github.com/MohamedAljoke/cka-sim/apps/go-shared => ../go-shared
