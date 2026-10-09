# Build cka-sim from zero

Instructions for rebuilding the simulator from an empty folder, as a list of **deliverables**.
A deliverable is one task that ends with something you can run and check. Each one depends
only on the ones before it, so finish them in order. The [study guide](STUDY-GUIDE.md) explains
how the finished system works. This file is the order to build it in.

Every deliverable has the same parts:

- **Goal**: what exists when you're done
- **Build**: the files you write, and what goes in them
- **Watch out**: what broke the first time, and why
- **Done when**: commands that prove it works. Don't move on until they pass.
- **Learn**: the concept the deliverable teaches

Commands marked **host** run in your WSL terminal. **node** means inside a node container
(`docker exec -it <node> bash`). **base** means inside the base container.

```
D1 kind by hand ─► D2 Calico ─► D3 node image ─► D4 base + ssh ─► D5 env in Go (cka-sim up)
                                                                      │
D6 task format ─► D7 scripts + runner ─► D8 grading ─► D9 selftest ◄──┘
                                                          │
D10 exam draw ─► D11 session + HTTP API ─► D12 panel ─► D13 exam/study ─► D14 status
                                                                              │
                                              D15 the task catalogue ◄────────┘
                                              D16 Makefile + CI
```

Milestones: after **D5** the environment builds itself. After **D9** you can practise one
task from the CLI. After **D13** you have the full timed exam.

## Conventions

These apply to every deliverable.

**Layout.** Each application lives in its own folder under `apps/`, for example `apps/cli` for
the `cka-sim` command and later `apps/web` for the web page. The repo root only holds what has
to be there: `README.md`, `docs/` and `.github/`.

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

---

## Phase A: The environment, by hand

### D0 · Prerequisites and an empty repo

- **Goal:** a Go module that builds, plus the tools every later step needs.
- **Build:**
  - Install Docker, kind (≥ 0.24), kubectl and Go (≥ 1.25).
  - `go mod init github.com/<you>/cka-sim`, an empty `cmd/cka-sim/main.go` that prints usage, and a `.gitignore` containing `bin/`.
- **Done when:**
  ```sh
  docker run --rm hello-world && kind version && kubectl version --client && go version   # host
  go build -o bin/cka-sim ./cmd/cka-sim && ./bin/cka-sim                                  # host
  ```
- [ ] done

### D1 · One kind cluster by hand

- **Goal:** a 1 control-plane + 1 worker cluster you created with the kind CLI, with no Go involved.
- **Build:** a scratch `kind.yaml`:
  ```yaml
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  nodes:
  - role: control-plane
  - role: worker
  kubeadmConfigPatches:
  - |
    kind: KubeletConfiguration
    failCgroupV1: false
  ```
  Then `kind create cluster --name test --image kindest/node:v1.35.8 --config kind.yaml`.
- **Watch out:** WSL2 runs cgroup **v1** by default, and Kubernetes 1.35's kubelet refuses to
  start on cgroup v1 unless `failCgroupV1: false` is set. Without the patch, `kind create`
  hangs and then fails at "Starting control-plane".
- **Done when:**
  ```sh
  kubectl --context kind-test get nodes                 # host: two nodes listed
  docker exec -it test-control-plane bash               # host
  ps -p 1 -o comm=; systemctl status kubelet            # node: PID 1 is systemd; the kubelet is a unit
  ls /etc/kubernetes/manifests                          # node: the static Pod manifests
  ```
- **Learn:** a kind node is a container running systemd, which runs containerd and the kubelet.
  On the control plane, the API server, etcd, scheduler and controller-manager are static Pods.
  It's a real kubeadm cluster; only the "machines" are containers. (Study guide step 2.)
- [ ] done

### D2 · Calico instead of kind's CNI

- **Goal:** pod networking that **actually enforces** NetworkPolicy on WSL2.
- **Build:** add to `kind.yaml`:
  ```yaml
  networking:
    disableDefaultCNI: true
    podSubnet: 192.168.0.0/16      # Calico's default pool, so its manifest needs no editing
  ```
  Recreate the cluster, then run:
  ```sh
  kubectl --context kind-test apply --server-side \
    -f https://raw.githubusercontent.com/projectcalico/calico/v3.32.2/manifests/calico.yaml
  ```
- **Watch out:**
  1. kindnet (kind's default CNI) enforces policies through nftables queues. The WSL2 kernel has
     no `CONFIG_NFT_QUEUE`, so policies **silently allow everything**. Calico uses iptables instead.
  2. calico-node mounts `/sys/kernel/security`, and WSL2 has no securityfs, so the pod never
     starts. Remove the volume and its mount with a strategic-merge patch (`$patch: delete`). The
     code has it as `noSecurityFS` in `internal/env/env.go`.
  3. Re-applying the manifest brings the mount back, so apply it only when the
     `calico-node` DaemonSet doesn't exist yet.
- **Done when:** `kubectl -n kube-system rollout status ds/calico-node` succeeds and every node is
  `Ready`. A default-deny NetworkPolicy then **blocks** `wget` between two busybox pods. Test that
  yourself; don't assume it works.
- **Learn:** without a CNI, nodes stay `NotReady`. NetworkPolicy is only an API object; whether
  it's enforced depends on the CNI and the kernel. (Study guide step 6.)
- [ ] done

### D3 · The node image: a node that looks like an exam host

- **Goal:** `cka-sim/node:v1.35.8`, a kind node image you can ssh into as `candidate`, with exam tooling.
- **Build:** `images/node/Dockerfile`, `FROM kindest/node:v1.35.8`:
  - `rm /etc/dpkg/dpkg.cfg.d/docker` (it strips man pages), then install `openssh-server sudo
    vim less man-db manpages bash-completion curl wget jq iputils-ping dnsutils netcat-openbsd`.
    Reinstall `coreutils grep sed tar systemd util-linux mount procps iproute2` so their man
    pages come back, then `mandb -q`.
  - Download `yq`, `etcdctl` and `etcdutl` into `/usr/local/bin`.
  - Add a `candidate` user with passwordless sudo, set `PasswordAuthentication no`, and `systemctl enable ssh`.
  - `images/node/exam-profile.sh` → `/etc/profile.d/cka-exam.sh`, also sourced from
    `/etc/bash.bashrc` (ssh login shells and `sudo -i` read different files). It sets up the `k` alias
    with completion and a prompt that shows the host name.
- **Watch out:** the base image has no sshd. You enable it, and systemd starts it when the
  node boots. Don't start it in `CMD`, because kind's entrypoint has to stay PID 1.
- **Done when:**
  ```sh
  docker build -t cka-sim/node:v1.35.8 images/node                       # host
  kind create cluster --name test --image cka-sim/node:v1.35.8 --config kind.yaml
  docker exec test-worker systemctl is-active ssh                         # host: active
  docker exec -it -u candidate test-worker bash -l -c 'type k; man -w systemctl; etcdctl version'
  ```
- **Learn:** extending someone else's image without breaking its entrypoint. Also, which shell
  startup files bash reads in which case. (Study guide step 3.)
- [ ] done

### D4 · The base host and the ssh trust

- **Goal:** a `cka-base` container with **no kubectl**, from which `ssh test-worker` drops you
  onto the node as `candidate`, where `k get nodes` works.
- **Build:**
  - `images/base/Dockerfile`: `ubuntu:24.04` with `openssh-client vim less bash-completion curl
    ca-certificates iputils-ping`. Remove the image's `ubuntu` user and add `candidate`.
    `CMD ["sleep","infinity"]` keeps the container alive.
  - `images/base/ssh_config` → `/etc/ssh/ssh_config.d/cka.conf`: `User candidate`,
    `StrictHostKeyChecking no`, `UserKnownHostsFile /dev/null`, `LogLevel ERROR`.
  - Wire them together by hand:
    ```sh
    docker run -d --name cka-base --hostname base --network kind cka-sim/base:latest
    docker exec -u candidate cka-base bash -c 'install -d -m 700 ~/.ssh && ssh-keygen -q -t ed25519 -N "" -f ~/.ssh/id_ed25519 && cat ~/.ssh/id_ed25519.pub' > pub
    docker exec test-control-plane cat /etc/kubernetes/admin.conf > admin.conf
    # for every node:
    docker exec -i test-worker bash -c 'install -d -o candidate -g candidate -m 700 /home/candidate/.kube /home/candidate/.ssh && install -o candidate -g candidate -m 600 /dev/stdin /home/candidate/.kube/config' < admin.conf
    docker exec -i test-worker install -o candidate -g candidate -m 600 /dev/stdin /home/candidate/.ssh/authorized_keys < pub
    ```
- **Watch out:** `--network kind` is what makes `ssh test-worker` resolve. Docker's embedded DNS
  only resolves container names for containers on the same user-defined network. Use
  `admin.conf` from **inside** the control plane: its server address is the in-network
  `https://<cluster>-control-plane:6443`, which works from any node. Your host kubeconfig points at a
  localhost port instead.
- **Done when:**
  ```sh
  docker exec -it -u candidate cka-base bash -l     # host
  kubectl get nodes                                 # base: command not found  ← intended
  ssh test-worker                                   # base: no prompt, no password
  k get nodes                                       # node: works
  ```
- **Learn:** Docker networks and DNS, how ssh key trust works, and why the same cluster needs two
  different kubeconfigs (study guide steps 4–5).
- [ ] done

---

## Phase B: Automate the environment in Go

### D5 · `cka-sim up / down / reset / shell`

- **Goal:** one command rebuilds everything from D1–D4: **two** clusters (`cka7491` with 2
  workers, `cka3962` with 1) plus base. Running it again on a half-built environment finishes it.
- **Build:**
  - `assets.go` at the module root: `//go:embed images tasks web` into `Assets embed.FS`. The binary carries
    its images, tasks and UI with it.
  - `internal/env/env.go`:
    - `New`: state directory `~/.cache/cka-sim`, plus an empty `docker/config.json` there.
      `Environ()` adds `DOCKER_CONFIG=<that dir>` to **every** docker/kind/kubectl call.
    - `Unpack`: write the embedded files to `~/.cache/cka-sim/assets` (docker build and bash need
      real files). Only rewrite files that changed, and write a temp file then rename it, so a
      running script never sees a half-written file.
    - `Up`: build both images → `kind get clusters` → create the missing clusters **in parallel**
      (`errgroup`), passing the config generated by `kindConfig(c)` on stdin → `installCNI` (D2,
      including the existence check and the patch) → `startBase` → `provisionNodes`.
    - `Down` (`docker rm -f cka-base` and `kind delete cluster` for each cluster), `Shell`
      (`docker exec -it -u candidate -w /home/candidate cka-base bash -l`), and `Ready`.
  - `cmd/cka-sim/main.go`: a `switch` on `args[0]` calling those, with Ctrl-C wired in through `signal.NotifyContext`.
- **Watch out:**
  - Docker Desktop on WSL often leaves a `credsStore` in `~/.docker/config.json` whose helper
    binary is broken, and then even public pulls fail. The private empty `DOCKER_CONFIG` avoids that.
  - If an `up` stops after creating a cluster but before Calico was ready, the next `up` must still
    run `installCNI` on that existing cluster.
  - `startBase` always deletes and recreates base, so the key is new each time. That's why
    `provisionNodes` must run after it, every time.
- **Done when:**
  ```sh
  go build -o bin/cka-sim ./cmd/cka-sim && bin/cka-sim up    # host: ~5 min the first time
  bin/cka-sim up                                             # host: again, fast, no errors
  bin/cka-sim shell   →  ssh cka3962-worker  →  k get nodes  # 2 nodes Ready
  bin/cka-sim down && bin/cka-sim up                         # starting from nothing works too
  ```
- **Learn:** driving CLIs from Go with `os/exec`, making commands safe to re-run (idempotence),
  running work concurrently with `errgroup`, and `embed`. (Study guide step 10.)
- [ ] done

---

## Phase C: Tasks and grading

### D6 · The task format and loader

- **Goal:** every task is a folder `tasks/<domain>/<id>/` that Go can load and validate.
- **Build:**
  - The format: `task.md` with frontmatter, then the question in Markdown:
    ```
    ---
    id: tr-kubelet
    title: Worker node NotReady
    domain: troubleshooting          # troubleshooting|architecture|networking|workloads|storage
    weight: 7                        # this task's share of the exam score
    cluster: cka3962
    host: cka3962-worker             # where the candidate must ssh
    order: last                      # optional, see D10
    ---
    ## Context ...
    ## Task ...
    ```
    Next to it: `setup.sh`, `check.sh`, `solution.sh` and `explain.md`.
  - `internal/tasks/tasks.go`: the `Domain` constants, `DomainWeights` (30/25/20/15/10, the CKA
    v1.35 split), `Parse` (a hand-written frontmatter parser, no YAML library),
    `Load(fs.FS)` (glob `tasks/*/*/task.md` and attach `explain.md`), and `Find`.
  - `tasks_test.go`: Parse accepts a good task and rejects a missing id/host/cluster/weight or an
    unknown domain. **Every shipped task has all five files.**
- **Done when:** `go test ./internal/tasks` passes, and a deliberately broken `task.md` makes it fail.
- **Learn:** keeping the content (tasks) separate from the engine (Go). Adding a task never
  needs Go changes. (Study guide step 7.)
- [ ] done

### D7 · Running task scripts: `lib.sh` and the runner

- **Goal:** `cka-sim practice tr-kubelet` breaks the worker's kubelet and prints the
  question. `cka-sim solution tr-kubelet -apply` fixes it.
- **Build:**
  - `tasks/lib.sh`, a set of helpers every script gets:
    `k` (= `kubectl --context "$CONTEXT"`), `on <node> <cmd>` and `on_host <cmd>` (= `docker exec -i <node> bash -c`),
    `check`, `eq`, `wait_for <seconds> <cmd>`, `fresh_ns`, `course_dir`, and `remember`/`recall`.
  - `internal/grader/grader.go`, `Runner.Script(ctx, task, name, timeout)`: runs
    `bash -c 'source "$LIB"; source "$TASK_DIR/$SCRIPT"'` **on your machine**, with
    `CLUSTER CONTEXT HOST TASK_ID TASK_DIR CKA_SIM_STATE` set, and stdout and stderr captured together.
  - Your first task, `tasks/troubleshooting/tr-kubelet/`: setup points the kubelet drop-in at
    `/usr/local/bin/kubelet` (which doesn't exist) and disables the unit. The solution puts the path back,
    then runs `daemon-reload` and `enable --now`.
  - The `list`, `practice`, `solution [-apply]` commands in main.go.
- **Watch out:**
  - The scripts run on the host, **not** on base and not over ssh: they need root on the nodes
    and kubectl from outside, and the candidate must not be able to see them.
  - `fresh_ns` force-deletes pods before deleting the namespace. A pod on a node whose kubelet
    another task broke can never confirm it has terminated, so the namespace would stay
    `Terminating` forever.
  - `remember` stores state on your machine (`~/.cache/cka-sim/state`), so nothing on an exam
    host gives away the answer.
- **Done when:** `practice tr-kubelet`, then inside a shell `ssh cka3962-worker`, then
  `systemctl status kubelet` shows `203/EXEC`. Fix it by hand, or run `solution -apply`, and the node goes back to `Ready`.
- **Learn:** the setup/check/solution pattern, and the difference between where the candidate works and where
  the grader runs.
- [ ] done

### D8 · Grading: a tiny text protocol with partial credit

- **Goal:** `cka-sim check tr-kubelet` prints a ✓/✗ line per check and `earned/total` points.
- **Build:**
  - In `lib.sh`, `check <points> <description> <cmd...>` prints `PASS <points> <description>` or
    `FAIL <points> <description>` and throws away the command's own output.
  - `grader.Parse` reads only lines that start with `PASS`/`FAIL` and ignores everything else, so
    scripts can print debugging output freely. It sums `Earned`/`Total`. `Grade` runs `check.sh` with a 3-minute timeout.
    `GradeAll` grades 4 tasks at a time.
  - `tr-kubelet/check.sh`: 4 points for the node being Ready, 2 for the unit being active, 2 for it being enabled (the "survives a reboot" part).
  - `grader_test.go`: mixed PASS/FAIL lines, noise lines in between, and empty output scoring 0/0.
- **Watch out:** a check needs to **wait** (`wait_for`) before deciding. A kubelet that was just fixed
  takes seconds before the node reports Ready.
- **Done when:** `check` scores 0/8 right after `practice` and 8/8 after `solution -apply`.
- **Learn:** a line-based protocol between bash and Go. It's simpler than JSON and good enough. (Study guide step 8.)
- [ ] done

### D9 · `selftest`: proving every task is fair

- **Goal:** `cka-sim selftest` runs setup → grade (**must be 0**) → solution → grade (**must be
  full marks**) for every task.
- **Build:** `selftest` in main.go. Tasks marked `order: last` run after the others. When a task
  fails, print both results so you can see which check is wrong.
- **Done when:** `bin/cka-sim selftest` shows ✓ for every task. From here on, run it after every new or changed task.
- **Learn:** a task's checks are tests too. If a task gives points for doing nothing, or its own
  solution can't reach full marks, the task is broken. (Study guide step 9.)
- [ ] done

---

## Phase D: The exam

### D10 · Drawing an exam that follows the curriculum

- **Goal:** `tasks.Draw(all, n, rng)` picks n tasks whose domain mix follows the curriculum weights, then shuffles them.
- **Build:** each domain's quota is `round(n × weight / 100)`, capped at the number of tasks that
  domain has. Fill any shortfall from the heaviest domains first, trim any excess from the lightest,
  then shuffle the result (the real exam doesn't group questions by domain).
  Tests: the domain split for n=16 matches the weights, and Draw never returns more tasks than exist.
- **Watch out:** `order: last` (the `Task.Last` field). An etcd snapshot-and-restore task rewinds every object
  created after its snapshot, so it has to be set up **after** every other task's objects exist.
  The other tasks set up 4 at a time; the `last` tasks run one by one afterwards.
- **Done when:** `go test ./internal/tasks -run Draw` passes.
- [ ] done

### D11 · Session store and HTTP API

- **Goal:** an HTTP server for the panel. If it restarts, it resumes the same exam.
- **Build:** `internal/server/server.go`:
  - `Exam{Study, StartedAt, Duration, TaskIDs, Flags, EndedAt, Results, Score}` saved as
    `~/.cache/cka-sim/exam.json` (write a temp file, then rename it). A `Store` with a mutex.
  - Routes (Go 1.22+ `ServeMux` patterns):
    `GET /api/exam`, `POST /api/flag/{id}`, `POST /api/end`,
    `POST /api/check/{id}` and `POST /api/reset/{id}` (both wrapped in `studyOnly`), and `GET /` serving the embedded `web/`.
  - `view()` builds what the browser sees. **In an exam, results, explanations and solutions stay
    hidden until the exam has ended.** In study mode they're always shown.
  - `Score`: each task's fraction of points earned, weighted by its `weight`, as a percentage. The pass mark is 66.
- **Watch out:**
  - Grade with `context.WithoutCancel(r.Context())`, so closing the tab doesn't cancel grading halfway.
  - A `grading` mutex means a double-click on "End exam" can't grade twice.
- **Done when:** `go test ./internal/server` passes (partial-credit scoring; study check reveals
  the solution; an exam refuses study actions with 403).
- **Learn:** keeping state in a file so a restarted process can carry on, and deciding on the server what the browser is allowed to see.
- [ ] done

### D12 · The panel

- **Goal:** `web/index.html`, **one file with no build step**: a question list, the current question with its
  "ssh <host>" box and copy buttons, flags, a countdown, End exam, and a results page.
  Study mode adds Check / Reset / Solution buttons.
- **Build:** plain JS: `api(path, method)` around `fetch`, a small `md()` renderer for task bodies,
  `render*()` functions, and `setInterval(tick, 1000)`. Time it against the server's `now`
  and `deadline`, not the browser clock.
- **Done when:** with an `exam.json` written by hand and the server from D11 running, the panel shows the questions
  and the clock counts down.
- **Learn:** how far you can get without a frontend framework. (Study guide step 11.)
- [ ] done

### D13 · `cka-sim exam` and `cka-sim study`

- **Goal:** the full flow: `exam [-n 16] [-minutes 120] [-port 8080] [-fresh=false] [-resume]`
  and `study [-fresh=false] [-resume] [id...]`.
- **Build:** in main.go:
  1. `listen(port)` **first**, so a port that's already taken fails right away instead of after 5 minutes of building clusters.
  2. `prepare(fresh)`: by default `down` + `up`, so nothing from an earlier attempt leaks into the new exam.
     `down` also deletes `exam.json` and the `state/` folder.
  3. `Draw` (exam) or `find` (study), then `setupAll`, then `store.Save`.
  4. `serve`: shut the HTTP server down cleanly on Ctrl-C. `-resume` skips steps 2–3.
- **Done when:** `bin/cka-sim exam -n 4 -minutes 15`, then work one task in `cka-sim shell`, then End exam.
  The results show that task scored and the untouched ones at 0. Stop the server with Ctrl-C, run
  `exam -resume`, and the same results come back.
- [ ] done

### D14 · `cka-sim status`

- **Goal:** one command that shows what the simulator is running and how much memory it uses.
- **Build:** `internal/env/status.go`. Find the node containers through kind's label
  `io.x-k8s.kind.cluster=<name>` and base by its name `^cka-base$`. Add up the memory from
  `docker stats --no-stream`, using `parseSize` for units like `GiB`/`MiB`/`kB` (with a unit test). Report the
  session state from `exam.json`.
- **Done when:** `bin/cka-sim status` lists 3 groups (two clusters and base), the total memory, and the session.
- [ ] done

---

## Phase E: Content and tooling

### D15 · The task catalogue: one deliverable per task

Write each task in the same order: **task.md → setup.sh → check.sh → solution.sh → explain.md → `selftest <id>` passes.**
Write the check before the solution, and make sure it scores 0 right after setup.

| Task | Domain | Host | What it trains | Special |
|---|---|---|---|---|
| `tr-kubelet` | troubleshooting | cka3962-worker | systemd unit, drop-in, `journalctl` | done in D7 |
| `tr-scheduler` | troubleshooting | cka3962-control-plane | broken static Pod manifest (typo in `kube-scheduler` command), Pod stuck Pending | breaks the control plane |
| `tr-service` | troubleshooting | cka7491-control-plane | Service selector/port vs Endpoints | |
| `ar-etcd` | architecture | cka3962-control-plane | `etcdctl snapshot save`, `etcdutl` restore | `order: last` |
| `ar-rbac` | architecture | cka7491-control-plane | Role/Binding, ServiceAccount, `auth can-i` | |
| `nw-netpol` | networking | cka7491-control-plane | ingress NetworkPolicy, pod + namespace selectors | needs Calico (D2) |
| `wl-rollout` | workloads | cka7491-control-plane | rollout history/undo, `maxUnavailable`/`maxSurge` | |
| `st-pvc` | storage | cka7491-control-plane | hostPath PV, PVC binding by storageClass, Pod mount | |

- **Watch out:** tasks that break a cluster (kubelet, scheduler, etcd) all live on `cka3962`, so
  they can't break the workload tasks on `cka7491`. Answer files go under `/opt/course/<id>/` on the host (`course_dir`).
- **Done when:** `bin/cka-sim selftest` passes for all 8 tasks.
- [ ] done

### D16 · Makefile and CI

- **Goal:** short commands for daily use, and CI that catches mistakes without needing Docker.
- **Build:**
  - `Makefile` (`make` alone prints help): `reset up down status exam study resume shell open`
    for daily use (`exam`/`study` pass `-fresh=false`, because `make reset` already built the
    environment), and `test check selftest` for development. `open` uses `explorer.exe` on WSL and `xdg-open` elsewhere.
  - `make check` = `go test -race ./...` + `gofmt -l` is empty + `go vet` + `bash -n` on every `.sh`.
  - `.github/workflows/ci.yml`: checkout, `setup-go` reading the Go version from `go.mod`, then `make check`.
    CI doesn't run `selftest`, because that needs Docker and real clusters. Run it locally.
- **Done when:** `make check` passes locally, and a push shows CI green.
- [ ] done

---

## Rebuild checklist (the whole thing, end to end)

```sh
make check                    # D6–D14 code is sound
make reset                    # D1–D5: environment from nothing
make status                   # D14
make selftest                 # D7–D9, D15: every task fair
make exam MINUTES=30 N=4      # D10–D13, then in another terminal: make open shell
```
