# cka-sim

A Kubernetes study environment anyone can install on their own machine.

Being rebuilt from scratch. The first implementation is kept on the
[`archive/v1`](https://github.com/MohamedAljoke/cka-sim/tree/archive/v1) branch for reference.

## Requirements

[Docker](https://docs.docker.com/get-started/get-docker/), running Linux containers. Run
`cka-sim doctor` to check your machine.

## Build from source

Needs Go 1.21 or newer. The `go` command downloads the exact toolchain pinned in
`apps/cli/go.mod` (Go 1.26.9) the first time you build.

```sh
cd apps/cli
go build -o bin/cka-sim ./cmd/cka-sim
./bin/cka-sim doctor      # check this machine
./bin/cka-sim up          # create the cluster: 1 control plane, 2 workers
./bin/cka-sim down        # delete it
```

## Docs

- [Build plan](docs/BUILD-FROM-ZERO.md): what's done, what's next, and the conventions.
- [Study guide](docs/STUDY-GUIDE.md): how each piece works, with things to run and questions to answer.

## Layout

| Path | What |
|---|---|
| `apps/cli` | the `cka-sim` command |
| `docs/` | the build plan and study guide |
