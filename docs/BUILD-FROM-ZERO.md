# Build cka-sim from zero

The build plan for cka-sim, written as a list of **deliverables**. A deliverable is one task that
ends with something you can run and check. Each one depends only on the ones before it, so finish
them in order. The [study guide](STUDY-GUIDE.md) explains how the pieces work. This file is the
order to build them in.

This is the **rebuild**. The first version lives on the
[`archive/v1`](https://github.com/MohamedAljoke/cka-sim/tree/archive/v1) branch. v1 worked, but
only on one machine: it needed kind, kubectl, Go and bash installed, and it built its images
locally. The rebuild has one goal on top of v1's: **anyone can download one binary and study,
with Docker as the only prerequisite.** Notes marked *(learned in v1)* are problems v1 already hit.

cka-sim is built for three kinds of people:

- **Practise only:** download a release binary, run it. Needs Docker and nothing else.
- **Customise:** clone the repo, change tasks or the page, build a binary, and delete the clone
  if they like. So the binary must never read the repo at runtime: everything it uses (the page,
  the tasks, the kind config) goes inside it.
- **Develop cka-sim:** run everything from source with fast reloads. This comes first; the
  release build and the embedding come later (D14, D17), and nothing in the dev setup may block them.

Every deliverable has the same parts:

- **Goal**: what exists when you're done
- **Build**: the files you write, and what goes in them
- **Watch out**: what broke, or will break, and why
- **Done when**: commands that prove it works. Don't move on until they pass.
- **Learn**: the concept the deliverable teaches

Commands marked **host** run in your terminal. **node** means inside a node container
(`docker exec -it <node> bash`). All Go commands run from `apps/cka-sim`.

```
✅ D0 skeleton + CI ─► ✅ D1 doctor ─► ✅ D2 up / down
                                            │
   ✅ D3 shell ─► ✅ D3b page + terminal ─► D4 task format ─► D5 runner + start ─► D6 grading + check ─► D7 selftest
                                                                              │
   D8 CNI ─► D9 node image ─► D10 base host + ssh ─► D11 status   ◄───────────┘
   (each pulled in when a task needs it)                │
                                                        ▼
   ✅ D12 exam draw ─► ✅ D13 session + API ─► ✅ D14 web panel ─► ✅ D15 exam from the page
                                                              │
                                  D16 task catalogue ◄────────┘
                                  D17 releases + install
```

Milestones: after **D2** anyone with Docker gets a real cluster. After **D7** you can practise
tasks from the CLI. After **D15** you have the full timed exam. After **D17** strangers can install it.

## Conventions

These apply to every deliverable.

**Layout.** Each application lives in its own folder under `apps/`: `apps/cka-sim` is the Go module
that builds the `cka-sim` binary, and `apps/web` is the browser page (Vite + TypeScript). The repo
root only holds what has to be there: `README.md`, `docs/` and `.github/`.

**Inside `apps/cka-sim`.** `internal/` holds the engine: checking the machine, the cluster, the
terminal, tasks, grading. Two thin front doors drive it: the CLI (`cmd/cli`: `main.go` is the entry point,
`cli.go` the commands) and the HTTP server the page talks to (`cmd/server`). They are separate
`main` packages: the server is not a CLI command. Anything a command can do,
a web request can trigger too. Neither front door holds logic of its own.

**Swappable parts.** Anything that talks to the outside world sits behind a small interface that
the engine owns: the cluster provider (kind today), how a shell is opened (`docker exec` today, ssh
to a base host from D10), where task scripts run. Swapping one means writing a new implementation
and changing the one line that picks it, not touching its callers. The same seam lets tests use a
fake, like `doctor.System` already does. Don't add an interface before there's a second
implementation or a test that needs it; do keep the outside-world calls in one package so adding
one later stays easy.

**Local development.** Go doesn't serve the page during development; Vite does.

```sh
cd apps/cka-sim
make dev              # cluster up, the Go backend, and the page on http://localhost:5173 (hot reload)
go run ./cmd/cli doctor   # any command, straight from source; `make` lists the shortcuts
```

The browser only ever talks to Vite. When the page needs the engine (the terminal first), a small
Go server answers on `127.0.0.1:7070` and Vite forwards `/ws` and `/api` to it, so the browser
sees one origin. The Go server only listens on localhost: its terminal is root on the cluster.

**Comments.** Write a comment only when it is really needed. About 90% of the time the code
should explain itself through clear names, small functions and a simple structure. Before you
write a comment, ask why the code needs it:

- **The code is confusing or badly written.** Refactor it instead: rename, extract a function,
  simplify the logic. A comment here only covers up the problem.
- **The code is clear, but something important can't be seen from reading it.** Then a short
  comment is justified. Examples: a workaround for a bug in a tool, an outside constraint
  (*"kind's docs recommend these inotify values"*), or a reason that isn't obvious (*"WSL2
  shares one kernel between distros, so the host limits apply"*).

A comment that repeats what the code says is noise. Say **why**, never **what**.

**Test files.** Put the tests first and the helpers (fakes, builders, small `find` or `parse`
functions) at the end of the file. Someone opening a test file wants to see what is being
tested. How the setup works comes second.

**Enums.** No `iota`. Give every constant an explicit value, and prefer string-typed constants
(`type Status string` with `OK Status = "ok"`). They read well in output and need no `String()` method.

**Docker is the only user prerequisite.** Anything a user would otherwise install (kind,
kubectl, bash, images) is either inside the binary or pulled by Docker. Check every new
deliverable against this.

---

## Phase A: Foundations ✅

### D0 · Skeleton and CI ✅

- **Goal:** a Go module under `apps/cka-sim` that builds a `cka-sim` binary for every OS, with CI proving it.
- **Build:**
  - `apps/cka-sim/go.mod` (`module github.com/MohamedAljoke/cka-sim/apps/cka-sim`), plus `.gitignore` with `bin/`.
  - `cmd/cli/main.go`: Ctrl-C turned into cancellation with `signal.NotifyContext`, and the version.
    `cmd/cli/cli.go`: a `switch` on `args[0]` and what each command calls. The version comes from `-ldflags "-X main.version=…"`
    in release builds, then `debug.ReadBuildInfo()`, then `dev`.
  - `.github/workflows/ci.yml`: vet + test on Ubuntu, macOS and Windows; gofmt; cross-compile
    linux/darwin/windows × amd64/arm64 with `CGO_ENABLED=0`.
- **Watch out:**
  - The module is not at the repo root, so `setup-go` needs both `go-version-file` and
    `cache-dependency-path` pointing into `apps/cka-sim`.
  - A nested module is versioned with **prefixed tags**: `apps/cka-sim/v0.1.0`, not `v0.1.0`. Go names
    a binary after its folder, so `go install …/apps/cka-sim/cmd/cli@latest` would install `cli`:
    releases (D17) are the supported install, and local builds always pass `-o bin/cka-sim`.
- **Done when:**
  ```sh
  go build -o bin/cka-sim ./cmd/cli && ./bin/cka-sim version    # host
  ```
  and CI is green on all three operating systems.
- **Learn:** Go modules in subdirectories, cross-compilation, build info.
- [x] done

### D1 · `cka-sim doctor` ✅

- **Goal:** one command that tells a new user whether their machine can run cka-sim, and exactly
  how to fix it when it can't.
- **Build:** `internal/doctor`:
  - `System` holds everything the checks read from the machine (`GOOS`, `LookPath`, `Run`,
    `ReadFile`). `Host()` is the real machine; tests pass fakes.
  - Checks, in order: docker installed → docker running (`docker info --format '{{json .}}'`) →
    Linux containers → cgroup version → CPUs ≥ 2 → memory ≥ 4 GiB → inotify limits (Linux only).
    If Docker is missing or not running, stop there, because nothing after it can be checked.
  - `Status` is `ok`, `warn` or `fail`. Only `fail` stops cka-sim. Every non-ok result has a
    `Fix` written for that platform: Docker Desktop settings, `.wslconfig`, `sysctl`, `usermod`.
- **Watch out:**
  - `docker info` still prints its JSON when the daemon is down, with the reason in `ServerErrors`.
  - `docker info` can hang while Docker Desktop is starting, hence the 20-second timeout.
  - CPU and memory come from `docker info`, not the host, because Docker Desktop runs inside a
    VM with its own limits.
  - Docker Desktop on Linux runs its own VM, so the host's inotify limits don't apply. WSL2
    distros share one kernel, so there they do.
- **Done when:**
  ```sh
  ./bin/cka-sim doctor          # host: every line [ok] (or [warn] with a fix)
  # stop Docker, run it again:  [fail] docker running, a fix line, exit code 1
  go test ./internal/doctor     # host
  ```
- **Learn:** inotify, cgroups, where Docker's resource limits live on each platform, and
  swapping the real machine for a fake in tests. (Study guide step 2.)
- [x] done

### D2 · `cka-sim up / down` with kind as a library ✅

- **Goal:** `cka-sim up` creates a real kubeadm cluster, 1 control plane and 2 workers, named
  `cka-sim`, with only Docker installed. `cka-sim down` deletes it.
- **Build:** `internal/cluster`:
  - kind as a Go library (`sigs.k8s.io/kind/pkg/cluster`), not the `kind` binary:
    `NewProvider(ProviderWithDocker(), ProviderWithLogger(kindcmd.NewLogger()))`.
  - The cluster config is a YAML string, the same format as `kind create cluster --config`.
  - `Create`: `CreateWithRawConfig`, `CreateWithKubeconfigPath(<UserConfigDir>/cka-sim/kubeconfig)`,
    `CreateWithWaitForReady(5m)`. `Exists` uses `List`. `Delete` removes the containers and the kubeconfig entry.
  - In `main.go`, `up` runs doctor first and prints only the problems. It stops on a `fail`, says
    "already up" when the cluster exists, and finishes by printing how to connect.
- **Watch out:**
  1. **cgroup v1.** Since Kubernetes 1.35 the kubelet refuses to start on cgroup v1
     (`failCgroupV1` defaults to `true`). WSL2 with Docker Desktop reports cgroup v1, and
     `kubeadm init` then fails at "Starting control-plane". `up` reads Docker's `CgroupVersion`
     and, on v1 only, adds a `kubeadmConfigPatches` entry with `kind: KubeletConfiguration` /
     `failCgroupV1: false`. Doctor warns and shows how to move to v2.
  2. When `Create` fails, kind deletes the nodes, so there are no logs left. To debug, add
     `CreateWithRetain(true)` for a while, then `docker exec cka-sim-control-plane journalctl -u kubelet`.
  3. kind's `Create` takes no context. Ctrl-C returns early and leaves containers behind, so the
     error message tells the user to run `cka-sim down`.
  4. Write the kubeconfig to our own file, never into the user's `~/.kube/config`.
  5. The kind version pins the Kubernetes version (kind v0.33.0 → `kindest/node:v1.37.0`).
     Bumping kind means running `up` again, not only `go get -u`.
- **Done when:**
  ```sh
  ./bin/cka-sim up                                               # host: ~2 min the first time
  docker exec cka-sim-control-plane kubectl get nodes            # host: 3 nodes Ready
  ./bin/cka-sim up                                               # host: "already up"
  ./bin/cka-sim down                                             # host
  docker ps --filter label=io.x-k8s.kind.cluster=cka-sim         # host: empty
  ```
- **Learn:** kind's create pipeline is `kubeadm init` and `kubeadm join`. Also cgroups, and
  kubeconfig files. (Study guide steps 3–5.)
- [x] done

---

## Phase B: The first practice loop

The goal of this phase is **get a task → work on it → check it**, on the plain kind cluster from
D2. Anything that needs Calico, a custom node image or the ssh setup waits for Phase C.

### D3 · `cka-sim shell`

- **Goal:** a terminal with `kubectl` working, without the user installing kubectl.
- **Build:** run `docker exec -it cka-sim-control-plane bash -l` from Go, with stdin, stdout and
  stderr attached. kind nodes already include kubectl and an admin kubeconfig.
- **Watch out:**
  - Pass `-t` only when stdin is a terminal, or piping into `cka-sim shell` breaks.
  - This is root on the control plane, not exam-like. D10 moves `shell` to a base host
    **without changing the command**.
- **Done when:** `./bin/cka-sim shell`, then `kubectl get nodes` lists 3 nodes, on Linux, macOS and Windows.
- [ ] done

### D3b · The page and its terminal (pulled forward from D14) ✅

- **Goal:** the exam screen without the exam: a question pane that says "No task yet" and a
  terminal into the cluster, in the browser. No timer, no tasks.
- **Build:**
  - `apps/web`: Vite + TypeScript, no framework. The two-pane exam layout. ✅
  - `apps/cka-sim/Makefile`: `make dev` brings the cluster up, then runs the server (`cmd/server`) and
    Vite together. ✅
  - `internal/terminal`: a shell on the cluster with a TTY, started at the browser's size,
    resizable, ended when the browser leaves. Behind an interface, so D10 can swap `docker exec`
    for ssh to the base host. ✅
  - `internal/server` + `cmd/server`: a Go server on `127.0.0.1:7070` with `/ws/terminal`,
    and a Vite proxy for `/ws`. ✅
  - xterm.js in the page, connected to that websocket. ✅
- **Watch out:**
  - v1 used creack/pty around the docker CLI, which has no Windows support. Use Docker's exec API
    with `Tty: true` instead: Docker makes the TTY, and resizing is an API call.
  - Start the shell only after the browser sends its size: a resize that arrives while the
    process starts can get lost, leaving it at 80x24. *(learned in v1)*
  - Closing the connection leaves bash running in the container. Tag each shell with an
    environment variable and kill everything carrying the tag (HUP, then KILL); sweep leftovers at
    startup. *(learned in v1)*
  - Refuse websocket connections from other origins, or any website could open a root shell
    through localhost. *(learned in v1)*
  - Binary messages are keystrokes and output; text messages are JSON control
    (`{"type":"resize","cols":…,"rows":…}`). Raise the read limit so a large paste fits.
  - xterm.css paints its viewport black around the themed text area; override it with the pane colour.
  - In WSL, a Windows browser reaches `localhost:5173` through WSL's port forwarding, so its
    connections show up in `ss` with no owning process. An open tab keeps one shell alive; that's
    not a leak.
  - Leave for later: several terminal tabs, reattaching after a reload, copy/paste shortcuts,
    pings to drop half-open connections (laptop sleep).
- **Done when:** `make dev`, open http://localhost:5173, and `kubectl get nodes` in the page lists 3 nodes.
- **Learn:** what a TTY is across a network, websockets, and keeping the outside world behind an interface.
- [x] done

### D4 · The task format and loader

- **Goal:** every task is a folder that ships inside the binary and that Go can load and validate.
- **Build:**
  - `apps/cka-sim/tasks/<id>/`: `task.md`, `setup.sh`, `check.sh`, `solution.sh`, `explain.md`.
    `task.md` has frontmatter, then the question in exam wording:
    ```
    ---
    id: wl-scale
    title: Scale a Deployment
    domain: workloads          # troubleshooting|architecture|networking|workloads|storage
    weight: 4                  # this task's share of the exam score
    ---
    ## Task ...
    ```
  - A small Go file inside `tasks/` with `//go:embed */*` exports the files as an `fs.FS`.
    `go:embed` can't reach parent folders, so the embed has to live next to the content.
  - `internal/tasks`: `Parse`, `Load(fs.FS)`, `Find`. The tests check that every shipped task has
    all five files and valid frontmatter.
- **Watch out:** v1's frontmatter had `cluster` and `host` because it had two clusters and ssh.
  Leave them out until D10 needs them.
- **Done when:** `go test ./internal/tasks` passes, and a broken `task.md` makes it fail.
- **Learn:** keeping the content (tasks) separate from the engine (Go). Adding a task never needs Go changes.
- [ ] done

### D5 · Running task scripts and `cka-sim start`

- **Goal:** `cka-sim tasks` lists the tasks. `cka-sim start <id>` runs `setup.sh` and prints the question.
- **Build:**
  - `lib.sh` helpers that every script gets: `k`, `on <node> <cmd>`, `wait_for <seconds> <cmd>`,
    `fresh_ns`, `remember`/`recall`.
  - A runner that executes `setup.sh`, `check.sh` and `solution.sh` with a timeout and captures their output.
- **Watch out:**
  - **Where do scripts run?** v1 ran them with bash **on the host**. That breaks "Docker only":
    Windows has no bash, and macOS ships bash 3.2. Run them **in a container** instead, for example
    a small runner container on the `kind` network with kubectl, bash and access to the nodes.
    Decide this before writing the runner.
  - The candidate must not be able to read the scripts or the `remember` state. *(learned in v1)*
  - `fresh_ns` force-deletes pods before deleting the namespace. A pod on a node whose kubelet
    another task broke never confirms termination, so the namespace stays `Terminating` forever. *(learned in v1)*
- **Done when:** `start <id>` breaks or prepares the cluster and prints the question. `solution <id> -apply` fixes it.
- **Learn:** the setup/check/solution pattern, and why the grader runs somewhere the candidate can't reach.
- [ ] done

### D6 · Grading and `cka-sim check`

- **Goal:** `cka-sim check <id>` prints a ✓/✗ line per check and `earned/total` points.
- **Build:**
  - `check <points> <description> <cmd...>` in `lib.sh` prints `PASS <points> <description>` or
    `FAIL <points> <description>` and throws away the command's own output.
  - `grader.Parse` reads only lines that start with `PASS`/`FAIL`, so scripts can print debug
    output freely. It sums earned and total points. Tests: mixed lines, noise between them, and empty output scoring 0/0.
- **Watch out:** *(learned in v1)*
  - A check must **wait** (`wait_for`) before deciding. A just-fixed kubelet takes seconds before the node is Ready.
  - Constraint checks only pass once the main goal does ("gating"). Otherwise doing nothing earns points.
  - Don't use `pipefail` in the runner: `grep -q` exits early and the command before it gets SIGPIPE.
- **Done when:** `check` scores 0 right after `start`, and full marks after `solution -apply`.
- **Learn:** a line-based protocol between bash and Go, the idea behind TAP. (Study guide step 8.)
- [ ] done

### D7 · `selftest` and the first tasks

- **Goal:** `cka-sim selftest` runs setup → grade (**must be 0**) → solution → grade (**must be
  full marks**) for every task, and two or three simple tasks pass it.
- **Build:** `selftest [id...]` in main.go. When a task fails, print both results. First tasks
  that work on plain kind: create and scale a Deployment, expose it with a Service, a
  ServiceAccount with RBAC.
- **Done when:** `./bin/cka-sim selftest` shows ✓ for every task. From here on, run it after every new or changed task.
- **Learn:** a task's checks are tests too. If a task gives points for doing nothing, or its own
  solution can't reach full marks, the task is broken. (Study guide step 9.)
- [ ] done

---

## Phase C: An exam-like environment

Pull each deliverable in when a task needs it, not before.

### D8 · A CNI that enforces NetworkPolicy

- **Goal:** NetworkPolicy tasks that grade on real traffic.
- **Build:** first **test kindnet** (kind's default CNI) on WSL2, macOS and Linux: does a
  default-deny policy block `wget` between two pods? If it does everywhere, skip the rest.
  Otherwise use Calico: `networking.disableDefaultCNI: true`, `podSubnet: 192.168.0.0/16`, and
  apply the Calico manifest from Go after `Create`.
- **Watch out:** *(learned in v1, re-check on kind v0.33)*
  1. kindnet enforced policies through nftables queues. The WSL2 kernel had no `CONFIG_NFT_QUEUE`,
     so policies **silently allowed everything**.
  2. calico-node mounts `/sys/kernel/security`, and WSL2 has no securityfs, so the pod never
     starts. Delete the volume and its mount with a strategic-merge patch (`$patch: delete`).
  3. Re-applying the manifest brings the mount back. Apply it only when `calico-node` doesn't exist yet.
- **Done when:** every node is `Ready`, and a default-deny policy **blocks** traffic. Test that; don't assume it.
- **Learn:** NetworkPolicy is only an API object. Whether it's enforced depends on the CNI and the
  kernel. (Study guide step 12.)
- [ ] done

### D9 · The node image, published so users never build it

- **Goal:** a node image with exam tooling, pulled from GHCR by `up`.
- **Build:** `images/node/Dockerfile` `FROM kindest/node:<version>`:
  - Remove `/etc/dpkg/dpkg.cfg.d/docker` (it strips man pages), then install `openssh-server sudo
    vim less man-db bash-completion curl jq dnsutils …` and run `mandb`. *(learned in v1)*
  - `etcdctl`/`etcdutl` pinned to the etcd version kubeadm uses; `yq`; a `candidate` user; the `k` alias with completion.
  - The `pkgs.k8s.io` apt repository, so `kubeadm upgrade` tasks can install the next version.
  - A GitHub Actions workflow builds it for amd64 and arm64 and pushes
    `ghcr.io/mohamedaljoke/cka-sim-node:<k8s version>`. `up` passes it with `CreateWithNodeImage`.
- **Watch out:** don't start sshd in `CMD`; enable the unit. kind's entrypoint has to stay PID 1. *(learned in v1)*
- **Done when:** `up` pulls the image, and `docker exec -it -u candidate cka-sim-worker bash -l -c 'type k; man -w systemctl; etcdctl version'` works.
- **Learn:** extending someone else's image without breaking its entrypoint; multi-arch images. (Study guide step 10.)
- [ ] done

### D10 · The base host and ssh

- **Goal:** `cka-sim shell` lands on a base container **without kubectl**, and `ssh cka-sim-worker`
  goes onto the node as `candidate`, where `k get nodes` works. This is the exam layout.
- **Build:** `images/base/Dockerfile` (also published to GHCR), started on the `kind` network. Its
  key is generated inside it and written to every node's `authorized_keys`, along with the
  in-network `admin.conf` as the node's kubeconfig. *(learned in v1)*
- **Watch out:**
  - `--network kind` is what makes `ssh cka-sim-worker` resolve. Docker's embedded DNS only
    resolves names on user-defined networks.
  - **One cluster or two:** v1 had two, so tasks that break a cluster (kubelet, scheduler, etcd)
    couldn't break the workload tasks. v2 keeps **one**. A task that breaks a node ships a `reset.sh`,
    which the engine runs (`tasks.Heal`) before each practice setup, before an exam's setups, after an
    exam's checks, and around each selftest. `order: last` keeps exam setups from colliding.
  - Base is a container we create ourselves, so it can carry labels (e.g. `com.docker.compose.project`)
    if we want it grouped in Docker Desktop. The kind nodes can't, because kind hardcodes their labels.
- **Done when:** `shell` → `kubectl` is not found → `ssh cka-sim-worker` with no password → `k get nodes` works.
- **Learn:** Docker networks and DNS, ssh key trust, and why one cluster needs two kubeconfigs. (Study guide step 11.)
- [ ] done

### D11 · `cka-sim status`

- **Goal:** one command that shows what cka-sim runs in Docker and how much memory it uses.
- **Build:** find containers by kind's label `io.x-k8s.kind.cluster=cka-sim` (plus base), and add
  up memory from `docker stats --no-stream` (`parseSize` handles `GiB`/`MiB`/`kB`, with a unit test).
- **Done when:** `status` lists every container with its role and state, plus the total memory.
- [ ] done

---

## Phase D: The exam

### D12 · Drawing an exam that follows the curriculum ✅

- **Goal:** `tasks.Draw(all, n, rng)` picks n tasks whose domain mix follows the CKA weights, then shuffles them.
- **Build:** largest-remainder quotas: each domain gets the whole part of `n × weight / 100`, and the
  leftover seats go to the biggest remainders. A quota is capped at the tasks the domain has, and any
  shortfall moves to the heaviest domains that still have tasks. Then shuffle (the real exam doesn't
  group by domain). Pass the RNG in, so tests can fix the seed.
- **Watch out:** an etcd snapshot-and-restore task rewinds every object created after its
  snapshot, so it must be set up **after** all the others. The same goes for tasks that break the
  scheduler or a kubelet. They carry `order: last` in their frontmatter, and the exam sets them up
  one by one after the rest. *(learned in v1)*
- **Done when:** `go test ./internal/tasks -run Draw` passes.
- [x] done

### D13 · Session store and HTTP API ✅

- **Goal:** the server runs exams. If it restarts, it resumes the same exam.
- **Build:** `internal/exam` owns the exam. Its state (tasks, setup status, flags, start, deadline, results)
  goes to `~/.local/state/cka-sim/exam.json` behind a small `Store` interface, written to a temp file then
  renamed. Routes on Go's `ServeMux` patterns (`POST /api/exam`, `PUT /api/exam/flags/{id}` …).
  **In an exam, results and solutions stay hidden until it ends.** Practice check and solution answer 409.
- **Watch out:** *(learned in v1)* grade with `context.WithoutCancel(r.Context())`, so closing
  the tab doesn't cancel grading halfway. One mutex covers every script, so a double-clicked "End exam"
  can't grade twice.
- **Done when:** `go test ./internal/exam ./internal/server` passes: weighted scoring, resume, and
  practice actions refused during an exam.
- **Learn:** keeping state in a file so a restarted process can carry on, and deciding on the server what the browser may see.
- [x] done

### D14 · The web panel ✅

- **Goal:** the question pills, the current question, flags, a countdown, End exam and results, beside
  the practice list (Check, Reset and Solution) and the page terminal.
- **Build:** the frontend lives in `apps/web` (`exam.ts`). The Go server lives in `apps/cka-sim`.
- **Watch out:** `go:embed` can't read outside `apps/cka-sim`, so the release build (D17) has to copy the
  web files into `apps/cka-sim` first. Time the countdown against the server's clock, not the browser's.
  *(learned in v1)*
- **Done when:** with a hand-written `exam.json`, the panel shows the questions and the clock counts down.
- [x] done

### D15 · Exam started from the page (no CLI) ✅

- **Goal:** the full timed exam end to end, started from the page's **Start exam** form. This replaced
  the planned `cka-sim exam` / `cka-sim study` commands, since the page is the UI.
- **Build:** `POST /api/exam` draws the tasks, sets them up, then starts the clock. There is no auto-end:
  after 0:00 the clock shows overtime. Ctrl-C on `make dev` shuts down cleanly, and the exam survives it.
- **Done when:** start an exam with 2 tasks and 5 minutes, solve one, restart the server, reload, End
  exam: the solved task scores full, the other 0, and the percentage matches the weights.
- [x] done

---

## Phase E: Content and release

### D16 · The task catalogue

Write each task in the same order: **task.md → setup.sh → check.sh → solution.sh → explain.md →
`selftest <id>` passes.** Write the check before the solution, and make sure it scores 0 right after setup.

| Task | Domain | What it trains | Needs |
|---|---|---|---|
| `tr-kubelet` | troubleshooting | systemd unit, drop-in, `journalctl` | D9 |
| `tr-scheduler` | troubleshooting | broken static Pod manifest, Pod stuck Pending | |
| `tr-service` | troubleshooting | Service selector/port vs Endpoints | |
| `ar-etcd` | architecture | `etcdctl snapshot save`, `etcdutl` restore | D9, `order: last` |
| `ar-rbac` | architecture | Role/Binding, ServiceAccount, `auth can-i` | |
| `ar-upgrade` | architecture | `kubeadm upgrade` of the control plane | D9 |
| `nw-netpol` | networking | ingress NetworkPolicy, pod and namespace selectors | D8 |
| `wl-rollout` | workloads | rollout history/undo, `maxUnavailable`/`maxSurge` | |
| `st-pvc` | storage | PV, PVC binding by storageClass, Pod mount | |

v1's versions of most of these are on `archive/v1` under `tasks/`.

- **Done when:** `selftest` passes for every task.
- [ ] done

### D17 · Releases and install

- **Goal:** a stranger installs cka-sim with one command.
- **Build:**
  - On a tag `apps/cka-sim/v*`, a workflow builds the six binaries with `-X main.version=…`, attaches
    them to a GitHub Release with checksums, and publishes the images from D9/D10.
  - An `install.sh` for `curl -fsSL … | sh` on Linux and macOS, a Homebrew tap, and a Windows download.
  - The README's quick start becomes: install → `cka-sim doctor` → `cka-sim up` → `cka-sim study`.
- **Done when:** on a machine with only Docker, the quick start works from a clean install.
- [ ] done

---

## Rebuild checklist (end to end)

```sh
cd apps/cka-sim
go vet ./... && go test ./...           # D0–D7, D12–D13 code is sound
go build -o bin/cka-sim ./cmd/cli
./bin/cka-sim doctor                    # D1
./bin/cka-sim up                        # D2: environment from nothing
./bin/cka-sim selftest                  # D5–D7, D16: every task fair
make dev                                # D12–D15: Start exam in the page
./bin/cka-sim down
```
