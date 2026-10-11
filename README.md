# cka-sim

A Kubernetes study environment anyone can install on their own machine.

Being rebuilt from scratch. The first implementation is kept on the
[`archive/v1`](https://github.com/MohamedAljoke/cka-sim/tree/archive/v1) branch for reference.

## Requirements

[Docker](https://docs.docker.com/get-started/get-docker/), running Linux containers. Run
`make doctor` to check your machine.

## Develop

Needs Docker, Go 1.21 or newer, and Node. The `go` command downloads the exact toolchain pinned in each
`go.mod` (Go 1.26.9) the first time you build. From the repo root, `make` lists the shortcuts:

```sh
make up               # the lab: a container that builds a 3-node cluster inside it (~2 min the first time)
make dev              # lab up, the Go backend, and the page with a terminal on http://localhost:5173
make selftest         # prove every task scores 0 before its solution and full after (needs the lab)
make shell            # a shell on the lab's base host, where the exam starts
make down             # delete the lab
```

The lab runs the same image here as on GKE, where each lab is a pod on its own VM. The backend reaches it
the same way in both places: through the agent inside it.

In the page, practise one task at a time, or press **Start exam** for a timed exam. Its tasks are drawn by
CKA domain weight, and it is scored at the end against the 66% pass mark.

## Docs

- [Build plan](docs/BUILD-FROM-ZERO.md): what's done, what's next, and the conventions.
- [Flows](docs/FLOWS.md): what happens, hop by hop, when you run a command or press a button.
- [Study guide](docs/STUDY-GUIDE.md): how each piece works, with things to run and questions to answer.

## Layout

| Path | What |
|---|---|
| `apps/platform` | what to run: the task catalog, exams, grading, the web server, and claiming labs |
| `apps/lab` | how it runs: the lab image (cluster, base host, agent); `api/` is all the platform imports |
| `apps/web` | the browser page (Vite); `npm run dev` serves it on :5173 |
| `infra/gke` | the GCP project and GKE cluster the hosted site runs on |
| `docs/` | the build plan, flows and study guide |
