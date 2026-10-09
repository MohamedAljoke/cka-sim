# cka-sim study guide

A step-by-step path through the simulator: first **understand** how it works, piece by piece,
then **deploy** it beyond your laptop. Each step has something to read, something to run,
and questions to answer before you move on. Tick the boxes as you go.

Have the environment up while you study (`make reset`, then `make status`). Commands marked
**host** run in your WSL terminal; **node** means inside a node container
(`docker exec -it <node> bash`); **base** means inside `make shell`.

---

## Part 1 — Understand it

### Step 1 · The big picture

- [ ] Read: the "How it differs from the real exam" section of the [README](../README.md).

There are three layers, each in the simplest tool that does the job:

| Layer | Tool | Lives in |
|---|---|---|
| Orchestration: build clusters, serve the panel, run grading | Go | `cmd/`, `internal/` |
| Exam content: break, check, fix | bash | `tasks/` |
| Exam panel | one HTML file, no build step | `web/index.html` |

Follow one action from end to end:

```
make exam  →  main.go exam()  →  listen on :8080
           →  env.Up / Ready  (docker + kind via os/exec)
           →  tasks.Draw      (pick tasks by domain weight)
           →  setupAll        (run each setup.sh, 4 at a time)
           →  store.Save      (~/.cache/cka-sim/exam.json)
           →  serve           (HTTP API + embedded panel)

"End exam" →  POST /api/end  →  grader.GradeAll  →  check.sh per task  →  PASS/FAIL lines  →  score
```

**Answer before moving on**
1. Why are the tasks in bash and not in Go?
2. Which process actually runs `setup.sh`: your machine, `base`, or a node?

---

### Step 2 · kind: Kubernetes nodes are containers

- [ ] Read: `kindConfig()` and `Up()` in `internal/env/env.go`.

kind runs every Kubernetes node as a Docker container that boots **systemd**. Inside it you
have containerd, a kubelet started by systemd, and, on control planes, the API server,
scheduler, controller-manager and etcd as **static Pods**. It's a real kubeadm cluster;
the "machines" are just containers.

Run:
```sh
docker ps --filter name=cka                       # host: one container per node
docker exec -it cka3962-control-plane bash        # host → node
ps -p 1 -o comm=                                  # node: PID 1 is systemd
systemctl status kubelet                          # node: kubelet is a systemd service
ls /etc/kubernetes/manifests                      # node: the static Pod manifests
crictl ps                                         # node: the containers the kubelet runs
cat /etc/systemd/system/kubelet.service.d/10-kubeadm.conf   # node: what tr-kubelet breaks
```

**Answer**
1. Who starts `kube-apiserver`: systemd, the kubelet, or kubeadm? (Look at the manifests directory.)
2. If you delete `/etc/kubernetes/manifests/kube-scheduler.yaml`, what happens and why?
3. Why can kind nodes run systemd and containerd at all? (Hint: `docker inspect cka3962-worker --format '{{.HostConfig.Privileged}}'`.)

---

### Step 3 · Making a node look like an exam host

- [ ] Read: `images/node/Dockerfile` and `images/node/exam-profile.sh`.

The image starts `FROM kindest/node:v1.35.8` and adds:
- **sshd**, a `candidate` user with passwordless sudo, and password login disabled
- **man pages:** the slim base deletes them through `/etc/dpkg/dpkg.cfg.d/docker`, so the
  image removes that file, reinstalls the packages and runs `mandb`
- **yq, etcdctl and etcdutl**, with etcd pinned to the version kubeadm 1.35 uses
- **the `k` alias and completion**, loaded from `/etc/bash.bashrc`

Run:
```sh
docker exec -it -u candidate cka3962-worker bash -l    # host
type k; man -w systemctl; etcdctl version              # node
```

**Answer**
1. Why does `ssh node "k get nodes"` fail when an interactive `ssh node` followed by `k get nodes` works?
   (Aliases only load in interactive shells.)
2. Why pin `ETCD_VERSION` instead of using the latest?

---

### Step 4 · Container networking: how `ssh cka3962-worker` works

- [ ] Read: `startBase()` and `provisionNodes()` in `internal/env/env.go`, plus `images/base/ssh_config`.

Every container joins the Docker network `kind`. On a user-defined Docker network, Docker
runs an **embedded DNS server (127.0.0.11)** that resolves container names to their IPs.
That's the whole trick: the container name is the hostname.

The ssh trust is wired like this:
1. `startBase` generates an ed25519 key **inside base**.
2. `provisionNodes` writes that public key to every node's `/home/candidate/.ssh/authorized_keys`.
3. `ssh_config` sets `User candidate` and skips host-key checks, because nodes are rebuilt all the time.

Run:
```sh
docker network inspect kind --format '{{range .Containers}}{{.Name}} {{.IPv4Address}}{{"\n"}}{{end}}'   # host
make shell                                              # host → base
cat /etc/resolv.conf                                    # base: nameserver 127.0.0.11
getent hosts cka3962-worker                             # base: name → IP via Docker DNS
ssh cka3962-worker                                      # base → node
```

**Answer**
1. Would this work on Docker's *default* `bridge` network? (No: it has no embedded DNS.)
2. Where does base's private key live, and why is it never copied out of base?
3. What would you change to make nested ssh (node → node) impossible, as on the exam?

---

### Step 5 · Kubeconfigs: who can talk to which cluster

- [ ] Read: the `admin.conf` part of `provisionNodes()`.

| Where | Kubeconfig | Why |
|---|---|---|
| Your machine | `~/.kube/config`, contexts `kind-cka7491` and `kind-cka3962` | written by `kind create cluster`; task scripts use these |
| Each node | `/home/candidate/.kube/config` = the cluster's `admin.conf` | so kubectl works after `ssh`, as on the exam |
| base | none, and no kubectl either | the exam's base host has no Kubernetes tools |

Run:
```sh
kubectl config get-contexts                                       # host
kubectl --context kind-cka3962 config view --minify | grep server # host: a 127.0.0.1 port mapped by Docker
ssh cka3962-worker 'grep server ~/.kube/config'                   # base: the in-network control-plane address
```

**Answer**
1. Why do the two `server:` addresses differ? (Port mapping to your host vs the Docker network.)
2. What's the security cost of giving `candidate` the admin kubeconfig? (Fine for an exam, never for production.)

---

### Step 6 · The CNI: Calico, and why not kind's default

- [ ] Read: `installCNI()` and `noSecurityFS` in `internal/env/env.go`.

kind ships its own CNI, **kindnet**. On your WSL kernel, kindnet's NetworkPolicy
enforcement needs `CONFIG_NFT_QUEUE`, which isn't there, so policies were *silently
allowed*. A NetworkPolicy task that grades on traffic would then be impossible. **Calico**
enforces policy with iptables and works.

Two settings make Calico fit kind:
- `disableDefaultCNI: true` and `podSubnet: 192.168.0.0/16`, which is Calico's default pool.
- A strategic-merge patch that deletes Calico's `/sys/kernel/security` mount, because WSL has
  no securityfs and the container wouldn't start. `installCNI` applies the manifest only
  once, since re-applying would bring that mount back.

Run:
```sh
kubectl --context kind-cka7491 -n kube-system get pods -l k8s-app=calico-node -o wide       # host
kubectl --context kind-cka7491 -n kube-system get ds calico-node -o yaml | grep -c kernel/security  # host: 0
docker exec cka7491-worker iptables-save | grep -c cali                                     # host: Calico's rules
```

**Answer**
1. Which component gives a Pod its IP: the kubelet, containerd, or the CNI plugin?
2. Why are nodes `NotReady` until the CNI is installed?
3. Why does a strategic-merge patch with `$patch: delete` work on lists that a JSON merge patch would replace?

---

### Step 7 · Anatomy of a task

- [ ] Read all five files of `tasks/troubleshooting/tr-service/`, then `tasks/lib.sh`.

```
task.md       frontmatter (id, domain, weight, cluster, host) + the question in exam wording
setup.sh      puts the cluster into its broken or starting state; must be re-runnable
check.sh      prints PASS/FAIL lines
solution.sh   the reference fix, with the diagnosis in comments
explain.md    the concept, shown after the exam
```

One important detail: **task scripts run on your machine, not over ssh.** `lib.sh` gives them:
- `k`: kubectl with the task's `--context`
- `on <node> <cmd>` / `on_host <cmd>`: `docker exec` into a node as root
- `check <points> <description> <command>`: runs the command and prints `PASS` or `FAIL`
- `wait_for <seconds> <command>`: retries until the command succeeds
- `fresh_ns <ns>`: force-deletes leftovers, then recreates the namespace
- `remember` / `recall`: state passed from setup to check (UIDs, generations), kept under
  `~/.cache/cka-sim/state` so it's invisible from the exam hosts

Run:
```sh
make study TASKS="tr-service"        # set it up, then in the panel:
                                     #   Check my work → read each check
                                     #   fix it via make shell → Check my work again
```

**Answer**
1. How does `tr-service`'s check prove you didn't "fix" it by editing the Deployment?
   (`remember generation` in setup; compared in check.)
2. Why does `fresh_ns` force-delete Pods before deleting the namespace?
   (A Pod on a node with a dead kubelet never confirms termination.)

---

### Step 8 · Grading: a tiny text protocol

- [ ] Read: `internal/grader/grader.go` and `grader_test.go`.

`check.sh` prints lines like `PASS 3 Service web has endpoints`. `grader.Parse` keeps only
lines that match `PASS|FAIL <int> <text>` and ignores the rest, so scripts can print
diagnostics freely. This is the idea behind TAP (Test Anything Protocol): agree on a line
format, and any language can be a test.

- **Partial credit:** score = Σ(task weight × earned/total) ÷ Σ weights. See `server.Score`.
- **Gating:** constraint checks only pass once the main goal does, otherwise doing nothing earns points.

**Answer**
1. Why was `pipefail` removed from the runner? (Hint: `grep -q` exits early → SIGPIPE upstream.)
2. What happens if `check.sh` hangs? (`Grade` uses a 3-minute `context.WithTimeout`.)

---

### Step 9 · The selftest: proving every task is fair

- [ ] Read: `selftest()` in `cmd/cka-sim/main.go`.

The rule for every task: **after setup it scores 0; after `solution.sh` it scores full marks.**

```sh
make selftest TASKS="tr-service"      # one task, about 1 minute
make selftest                         # all of them, about 12 minutes
```

The bugs it caught:
- **Free points:** constraint checks passed on untouched tasks, which led to gating.
- **Incomplete etcd restore:** restoring etcd under a running API server left it serving
  stale objects from its watch cache. The fix: stop the API server, restore with
  `--bump-revision --mark-compacted`, then start it again.
- **Interfering tasks:** a Pod landed on the node another task breaks, which jammed the
  namespace. The fix is a `nodeSelector`, plus `fresh_ns` force-deleting Pods.

**Answer**
1. Why must `ar-etcd` run last (`order: last`)? What would a restore do to tasks set up after the snapshot?
2. Write one invariant like this for a project of your own.

---

### Step 10 · The Go program

The code follows the [conventions](BUILD-FROM-ZERO.md#conventions): it should explain itself,
so a comment means there is something you can't see from the code alone. When you find a
comment, ask what it tells you that the code doesn't.

Read in this order, about 15 minutes each:

| File | Concepts to look for |
|---|---|
| `assets.go` + `env.Unpack()` | `//go:embed` puts everything in one binary; unpacking writes only changed files, via **write-temp-then-rename**, which is atomic |
| `internal/grader/grader.go` | `exec.CommandContext` + `context.WithTimeout`; env vars as the script interface; `errgroup.SetLimit(4)`; `slices.Clip` against shared-backing-array appends |
| `internal/tasks/tasks.go` | frontmatter parsing; `Draw` uses domain quotas then shuffles, with the RNG passed in so tests can fix the seed |
| `internal/server/server.go` | Go 1.22 routing (`"POST /api/flag/{id}"`, `r.PathValue`); `studyOnly` as hand-written middleware; a mutex so grading never runs twice; `context.WithoutCancel` so a closed tab doesn't cancel grading |
| `internal/env/status.go` | finding containers by label (`io.x-k8s.kind.cluster`); parsing `docker stats` output |
| `cmd/cka-sim/main.go` | one `flag.FlagSet` per subcommand; `signal.NotifyContext` turns Ctrl-C into cancellation; `listen()` claims the port **before** slow setup, so it fails fast |

Run:
```sh
make check                              # gofmt, vet, go test -race, bash -n
go test -race -run Study ./internal/server -v
```

**Answer**
1. Why is study mode enforced on the **server** and not just hidden in the UI?
2. What breaks if two goroutines `append` to the same `Environ` slice without `slices.Clip`?
3. Why is `exam.json` written via a temp file and `os.Rename`, not `os.WriteFile` directly?

---

### Step 11 · The panel

- [ ] Read: `web/index.html` (the `<script>` part).

- **No framework:** `render()` rebuilds the main area from `state.exam`, and every action calls the
  API and re-renders with its response.
- **Timer:** the server sends `now`; the page stores the offset to its own clock and counts
  locally, so a wrong browser clock can't change your deadline.
- **Persistence:** every bit of state is in `exam.json`, so the browser can close any time and
  `make resume` restores the panel.

**Answer**
1. In an exam, where would a cheater look for solutions, and why don't they find them?
   (Check `view()`: solutions are only filled in when `Ended || Study`.)

---

### Step 12 · Environment quirks: why they happened

| Symptom | Root cause | Fix in the code |
|---|---|---|
| kubelet refused to start | WSL uses cgroup v1; K8s 1.35 refuses it unless told otherwise | `failCgroupV1: false` in `kindConfig` |
| NetworkPolicy allowed everything | kindnet needs `CONFIG_NFT_QUEUE` | Calico (Step 6) |
| calico-node never started | no securityfs on WSL | `noSecurityFS` patch |
| image pulls failed | Docker Desktop credential helper left in `~/.docker/config.json` | a private, empty `DOCKER_CONFIG` |

Run: `stat -fc %T /sys/fs/cgroup` prints `tmpfs` on cgroup v1 and `cgroup2fs` on v2.

---

### Checkpoint: explain these out loud

- [ ] How `ssh cka3962-worker` reaches the right container, starting from DNS.
- [ ] What a static Pod is, and who restarts it when you edit its manifest.
- [ ] Why NetworkPolicy depends on the CNI.
- [ ] Why the etcd restore needs the API server stopped.
- [ ] How a `check.sh` line turns into a percentage on the results page.
- [ ] One concurrency bug this codebase avoids, and how.

If you can do all six, you understand the simulator. The best next step is **writing a task
yourself** (CoreDNS troubleshooting is a good first one) and proving it with `make selftest`.

---

## Part 2 — Deploy it

The goal: use the simulator from any browser, not just your laptop. Do the phases in order;
each one works on its own.

> ⚠ **Security model, read this first.** kind nodes are **privileged** containers. A shell on
> one is effectively **root on the machine that runs Docker**. That's fine when the machine
> is yours and only you use it. It is **not** fine for strangers sharing one Docker host.
> Phases 1–4 are single-user for this reason; Phase 5 is what multi-user takes.

### Phase 1 · One cloud VM, by hand

- [ ] Create a Linux VM: Ubuntu 24.04, at least **4 vCPU / 8 GB RAM**, with 16 GB to be
      comfortable; about 40 GB disk. Most providers have one in this range; the cost adds up
      only while it runs.
- [ ] Install Docker, kind, kubectl and Go, then clone the repo.
- [ ] `make reset && make status`.

What's different from WSL: a cloud VM normally runs **cgroup v2** and has securityfs, so
two of the three WSL workarounds aren't needed there. They're harmless, so leave them.

**Learn:** VM sizing, package installs, how kind behaves on a "real" Linux host.

### Phase 2 · Reach the panel safely

The panel binds to `127.0.0.1` on purpose. Don't change that. Start with an **ssh tunnel**:

```sh
ssh -L 8080:127.0.0.1:8080 you@your-vm      # then open http://localhost:8080 on your laptop
```

Once that works, move to a real URL:
- [ ] A domain name pointing at the VM.
- [ ] A reverse proxy (Caddy is the simplest, with automatic HTTPS from Let's Encrypt) in
      front of `127.0.0.1:8080`.
- [ ] **Authentication** at the proxy (basic auth at minimum). The panel has no login of its own,
      and it includes a shell on base, so without authentication anyone who finds the URL gets one.
- [ ] A firewall: allow only 22 and 443.

**Learn:** port forwarding, reverse proxies, TLS certificates, why services bind to localhost.

### Phase 3 · A terminal in the browser

The panel has a terminal built in: xterm.js in the page, a websocket endpoint in the Go server
(`internal/terminal`), and a pty running `docker exec -it -u candidate cka-base bash -l`. So
once Phase 2 puts the panel behind your proxy and authentication, the browser is all you need.
Read `internal/terminal/terminal.go` to see how it works. Alternatively, a standalone web terminal:
- [ ] Run a web terminal such as **ttyd** with the command
      `docker exec -it -u candidate cka-base bash -l`, bound to localhost.
- [ ] Put it behind the same proxy and authentication, for example at `/terminal`.
- [ ] Make sure the proxy forwards websockets (`Upgrade` headers). Caddy does this by
      default. The panel's own terminal needs it too.

**Learn:** websockets, pseudo-terminals (pty), how browser terminals work.

### Phase 4 · Automate the VM

- [ ] **cloud-init** (or a shell script) that installs everything and runs `make reset` on first boot.
- [ ] **Terraform** (or OpenTofu) for the VM, DNS record and firewall, so `terraform apply`
      creates it all and `terraform destroy` removes it.
- [ ] Cost control: destroy the VM when you aren't studying, or stop it on a schedule.

**Learn:** infrastructure as code, immutable infrastructure, cost awareness.

### Phase 5 (optional) · Many users: a real platform

This is what killer.sh or KodeKloud do, and where most of the engineering goes:

| Problem | Direction |
|---|---|
| **Isolation:** privileged containers mean root on the host | one environment per user, inside its own VM: short-lived cloud VMs, Firecracker microVMs, or KubeVirt |
| **Lifecycle:** create on start, destroy after 2 hours | a scheduler or queue; idle and TTL cleanup |
| **Cost:** about 5 GB of RAM per session | session limits, pre-warmed pools, shorter exams |
| **Accounts:** who is who | real authentication; exam state per user instead of one `exam.json` |
| **Running the platform on Kubernetes** | an **operator**: a CRD `ExamSession`, and a controller that creates and cleans up each session's VM |

**Learn:** multi-tenancy, sandboxing, controllers and CRDs (the deepest Kubernetes topic).

---

### Deployment checkpoint

- [ ] I can reach the panel from another device over HTTPS, with a login.
- [ ] I can work a task entirely in the browser.
- [ ] I can destroy and recreate the whole VM with one command.
- [ ] I can explain why Phase 5 needs a VM per user, and not a container per user.
