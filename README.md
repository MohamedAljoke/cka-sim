# cka-sim

A Kubernetes study environment anyone can install on their own machine.

Being rebuilt from scratch. The first implementation is kept on the
[`archive/v1`](https://github.com/MohamedAljoke/cka-sim/tree/archive/v1) branch for reference.

## Requirements

[Docker](https://docs.docker.com/get-started/get-docker/), running Linux containers. Run
`cka-sim doctor` to check your machine.

## Build from source

Needs Go 1.21 or newer. The `go` command downloads the exact toolchain pinned in
`apps/cka-sim/go.mod` (Go 1.26.9) the first time you build.

```sh
cd apps/cka-sim
go build -o bin/cka-sim ./cmd/cli
./bin/cka-sim doctor      # check this machine
./bin/cka-sim up          # create the cluster: 1 control plane, 2 workers
./bin/cka-sim down        # delete it
```

## Develop

Needs Go and Node. From the repo root, `make` lists the shortcuts:

```sh
make dev              # cluster up, the Go backend, and the page with a terminal on http://localhost:5173
make selftest         # prove every task scores 0 before its solution and full after (needs the cluster)
make doctor           # or any other CLI command, straight from source
```

In the page, practise one task at a time, or press **Start exam** for a timed exam. Its tasks are drawn by
CKA domain weight, and it is scored at the end against the 66% pass mark.

## Docs

- [Build plan](docs/BUILD-FROM-ZERO.md): what's done, what's next, and the conventions.
- [Flows](docs/FLOWS.md): what happens, hop by hop, when you run a command or press a button.
- [Study guide](docs/STUDY-GUIDE.md): how each piece works, with things to run and questions to answer.

## Layout

| Path | What |
|---|---|
| `apps/cka-sim` | the `cka-sim` binary: CLI, web server and the engine they share |
| `apps/web` | the browser page (Vite); `npm run dev` serves it on :5173 |
| `docs/` | the build plan and study guide |
