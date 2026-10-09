# cka-sim

A Kubernetes study environment anyone can install on their own machine.

Being rebuilt from scratch. The first implementation is kept on the
[`archive/v1`](https://github.com/MohamedAljoke/cka-sim/tree/archive/v1) branch for reference.

## Requirements

[Docker](https://docs.docker.com/get-started/get-docker/), running Linux containers. Run
`cka-sim doctor` to check your machine.

## Build from source

Needs Go 1.25 or newer.

```sh
cd apps/cli
go build -o bin/cka-sim ./cmd/cka-sim
./bin/cka-sim doctor
```

## Layout

| Path | What |
|---|---|
| `apps/cli` | the `cka-sim` command |
| `docs/` | the build plan and study guide |
