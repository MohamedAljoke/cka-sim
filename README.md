# cka-sim

A Kubernetes study environment anyone can install on their own machine.

Being rebuilt from scratch. The first implementation is kept on the
[`archive/v1`](https://github.com/MohamedAljoke/cka-sim/tree/archive/v1) branch for reference.

You get a real 3-node Kubernetes cluster in your browser, with a terminal on an exam-style base host. Practise
CKA tasks one at a time and press **Check** to grade them, or take a timed mock exam scored against the 66%
pass mark.

## Try it

You need [Docker](https://docs.docker.com/get-started/get-docker/) running Linux containers, with at least
4 GB of memory and 2 CPUs for it and about 7 GB of free disk. On Windows, run this inside WSL2.

```sh
curl -fsSL https://raw.githubusercontent.com/MohamedAljoke/cka-sim/main/infra/local/install.sh | sh
```

That installs the `cka-sim` command and starts it. The first start downloads the images and builds the
cluster, which takes 5 to 10 minutes. Then open <http://localhost:7080> and log in as `dev@cka-sim.local`
with the password `cka-sim-dev`.

```sh
cka-sim down        # stop it; your progress and the downloaded images stay
cka-sim up          # start it again (under a minute after the first time)
cka-sim status      # what is running
cka-sim logs lab    # follow one part's logs: lab, app, auth, mcp or postgres
cka-sim shell       # a shell on the base host, where the exam starts
cka-sim update      # get the latest version
cka-sim uninstall   # delete it and everything it stored
```

Prefer to build it yourself? You need only Docker and `make`:

```sh
git clone https://github.com/MohamedAljoke/cka-sim && cd cka-sim && make try
```

**If something goes wrong:**
- **`cka-sim: command not found`:** add `~/.local/bin` to your `PATH`, or run `~/.cka-sim/cka-sim`.
- **Port 7080 is taken:** run `CKA_SIM_PORT=8090 cka-sim up`, and set it the same way for every `up`.
- **"Pool overlaps with other one on this space":** another Docker network already uses `10.237.73.0/24`.
  Remove it, or change the subnet and the two addresses in `~/.cka-sim/compose.yml`.
- **The first start times out:** run `cka-sim logs lab` to see how far the cluster got, then `cka-sim up`
  again. It picks up where it left off.

The Assistant panel runs Claude Code in the lab; log in to Claude there the first time you open it.

## Develop

Needs Docker, Go 1.21 or newer, and Node. The `go` command downloads the exact toolchain pinned in each
`go.mod` (Go 1.26.9) the first time you build. From the repo root, `make` lists the shortcuts:

```sh
make up               # the lab: a container that builds a 3-node cluster inside it (~2 min the first time)
make dev              # lab up, the Go backend, and the page with a terminal on http://localhost:5173
make selftest         # prove every task scores 0 before its solution and full after (needs the lab)
make shell            # a shell on the lab's base host, where the exam starts
make down             # stop it; the cluster, accounts and images stay
make try              # build the images from source and run them the way `cka-sim up` does
```

`make dev`, `make try` and `cka-sim up` all run the Compose project in `infra/local/compose.yml`, so they share
one lab, one database and one set of secrets. `make dev` starts only the lab and Postgres from it and runs the
rest from source on this machine.

Run `make doctor` to check your machine.

The lab runs the same image here as on GKE, where each lab is a pod on its own VM. The backend reaches it
the same way in both places: through the agent inside it.

Exam tasks are drawn by CKA domain weight. Images for `cka-sim up` are published to GHCR by
`.github/workflows/images.yml` on every push to `main`.

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
| `infra/local` | `cka-sim` on one machine with only Docker: the Compose file, the command and its installer |
| `docs/` | the build plan, flows and study guide |
