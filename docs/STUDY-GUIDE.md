# cka-sim study guide

A step-by-step path through cka-sim: first **understand** how it works, piece by piece, then
**deploy** it beyond your laptop. Each step has something to read, something to run, and questions
to answer before you move on. Tick the boxes as you go.

The guide grows with the rebuild ([build plan](BUILD-FROM-ZERO.md)). Steps 1–6 cover code that
exists now. Steps 7–13 are **notes from v1**: the ideas still hold, but their file paths and
commands refer to the [`archive/v1`](https://github.com/MohamedAljoke/cka-sim/tree/archive/v1)
branch. Each one says which deliverable will rewrite it.

Have the cluster up while you study (`cd apps/cka-sim && ./bin/cka-sim up`). Commands marked
**host** run in your terminal; **node** means inside a node container (`docker exec -it <node> bash`).

---

## Part 1 — Understand it

### Step 1 · The big picture

- [ ] Read: the [README](../README.md) and the top of the [build plan](BUILD-FROM-ZERO.md).

The goal: **anyone downloads one binary and studies, with Docker as the only prerequisite.**
Everything else is either inside the binary or pulled by Docker.

| Layer | Tool | Lives in | Status |
|---|---|---|---|
| Orchestration: check the machine, build the cluster, grade | Go | `apps/cka-sim` | doctor, up, down |
| Kubernetes nodes | kind, used as a Go library | `apps/cka-sim/internal/cluster` | done |
| Exam content: break, check, fix | bash | `apps/cka-sim/tasks` | D4–D7 |
| Exam panel | web | `apps/web` | D14 |

Follow `cka-sim up` from end to end:

```
cka-sim up  →  main.go runUp()
            →  doctor.Run         (is Docker installed, running, big enough?)
            →  cluster.Exists     (kind: List the clusters)
            →  cluster.Create     (read Docker's cgroup version → kind config YAML)
                 →  kind: Provision nodes → kubeadm init → CNI → StorageClass → kubeadm join → wait Ready
            →  print how to connect
```

**Answer before moving on**
1. Why use kind as a **library** instead of running the `kind` command like v1 did?
2. What does a user need installed to run `cka-sim up`? What did they need for v1?

---

### Step 2 · `doctor`: checking someone else's machine

- [ ] Read: `apps/cka-sim/internal/doctor/doctor.go` and its test.

You can't see the machines your users run cka-sim on, so doctor checks them for you. Each check
returns a `Result` with a status (`ok`, `warn`, `fail`), a detail, and a **fix written for that
platform**. Only `fail` stops cka-sim.

| Check | Why it matters |
|---|---|
| docker installed / running | without Docker there are no nodes |
| Linux containers | Docker Desktop on Windows can be switched to Windows containers; kind needs Linux |
| cgroup version | the kubelet refuses cgroup v1 since Kubernetes 1.35 (step 4) |
| CPUs, memory | read from `docker info`, not the host, because Docker Desktop runs in a VM with its own limits |
| inotify limits | every node's systemd, kubelet and containerd watch files; low limits cause `too many open files` |

`System` holds every way doctor touches the machine (`LookPath`, `Run`, `ReadFile`, `GOOS`). The
real one is `Host()`; tests pass fakes, so they can pretend to be a Mac with Docker Desktop
stopped, or WSL2 with 2 GiB of memory.

Run:
```sh
./bin/cka-sim doctor                                            # host
docker info --format '{{.CgroupVersion}} {{.NCPU}} {{.MemTotal}} {{.OperatingSystem}}'   # host
cat /proc/sys/fs/inotify/max_user_watches /proc/sys/fs/inotify/max_user_instances      # host (Linux/WSL)
```

**Answer**
1. What's the difference between an inotify **instance** and a **watch**? Which one does `max_user_instances` limit?
2. Why is inotify skipped for Docker Desktop on Linux, but checked for Docker Desktop on WSL2?
3. Why does doctor stop after "docker running" fails instead of running the other checks?

---

### Step 3 · kind: Kubernetes nodes are containers

- [ ] Read: `apps/cka-sim/internal/cluster/cluster.go`, then kind's `pkg/cluster/internal/create/create.go`
      in your module cache (`go list -m -f '{{.Dir}}' sigs.k8s.io/kind`).

kind runs every Kubernetes node as a Docker container that boots **systemd**. Inside it there is
containerd, a kubelet started by systemd, and on the control plane the API server, scheduler,
controller-manager and etcd as **static Pods**. It's a real kubeadm cluster; only the "machines"
are containers.

`Create` runs a pipeline of actions. Each one is a line of `up`'s output:

| Action | Output | What happens |
|---|---|---|
| provision | `Preparing nodes 📦 📦 📦` | pull `kindest/node`, start 3 privileged containers on the `kind` network |
| config | `Writing configuration 📜` | generate `kubeadm.conf` per node and apply `kubeadmConfigPatches` |
| kubeadminit | `Starting control-plane 🕹️` | `kubeadm init` on the control plane |
| installcni | `Installing CNI 🔌` | kindnet, kind's pod network |
| installstorage | `Installing StorageClass 💾` | local-path-provisioner as the default StorageClass |
| kubeadmjoin | `Joining worker nodes 🚜` | `kubeadm join` on each worker |
| waitforready | `Waiting ≤ 5m0s …` | wait for the control plane to be Ready |

Run:
```sh
docker ps --filter label=io.x-k8s.kind.cluster=cka-sim     # host: one container per node
docker exec -it cka-sim-control-plane bash                 # host → node
ps -p 1 -o comm=                                           # node: PID 1 is systemd
systemctl status kubelet                                   # node: the kubelet is a systemd service
ls /etc/kubernetes/manifests                               # node: the static Pod manifests
crictl ps                                                  # node: the containers the kubelet runs
cat /kind/kubeadm.conf                                     # node: the config kind generated
```

**Answer**
1. Who starts `kube-apiserver`: systemd, the kubelet, or kubeadm? (Look at the manifests directory.)
2. If you delete `/etc/kubernetes/manifests/kube-scheduler.yaml`, what happens and why?
3. Why can kind nodes run systemd and containerd at all?
   (Hint: `docker inspect cka-sim-worker --format '{{.HostConfig.Privileged}}'`.)
4. Why did we choose kind and not minikube or k3s? (Hint: what does a CKA candidate need to see on a node?)

---

### Step 4 · cgroups, and why `up` failed the first time

- [ ] Read: `kindConfig` and `allowCgroupV1Patch` in `cluster.go`, and `checkCgroupVersion` in `doctor.go`.

**cgroups** (control groups) are the Linux kernel feature that puts processes into groups and
limits or measures what each group uses: CPU, memory, number of processes. Containers are built on
them. When a pod has `resources.limits.memory: 256Mi`, the kubelet creates a cgroup with that cap,
and a container that goes over it gets `OOMKilled`.

**v1 and v2** are two versions of the kernel interface, not two features. v1 has a separate
hierarchy per resource (`/sys/fs/cgroup/memory`, `/sys/fs/cgroup/cpu`, …). v2 has one unified
hierarchy and better memory handling. Pod limits work on both.

Since Kubernetes **1.35**, the kubelet refuses to start on a cgroup v1 host (`failCgroupV1`
defaults to `true`). WSL2 with Docker Desktop reports cgroup v1, so `kubeadm init` failed. `up` now
reads Docker's cgroup version and, on v1 only, adds this to the kind config:

```yaml
kubeadmConfigPatches:
  - |
    kind: KubeletConfiguration
    failCgroupV1: false
```

kind merges that snippet into the kubelet's configuration. It's the same mechanism you'd use on
the exam to change a kubeadm-managed setting.

Run:
```sh
docker info --format '{{.CgroupVersion}}'                       # host: 1 or 2
stat -fc %T /sys/fs/cgroup                                      # host: tmpfs = v1, cgroup2fs = v2
docker exec cka-sim-worker cat /var/lib/kubelet/config.yaml | grep -i cgroup   # host
docker exec cka-sim-worker ls /sys/fs/cgroup | grep kubepods    # host: where pods' cgroups live
```

**Answer**
1. Does `failCgroupV1: false` change how pod limits are enforced? (No: it only lets the kubelet start.)
2. On WSL2, how do you move to cgroup v2? (`kernelCommandLine = cgroup_no_v1=all` in `.wslconfig`, then `wsl --shutdown`.)
3. Which QoS class does a pod get with requests equal to limits? With no requests at all?

---

### Step 5 · Kubeconfigs: who can talk to the cluster

- [ ] Read: `New` and `Create` in `cluster.go` (`KubeconfigPath`, `CreateWithKubeconfigPath`).

| Where | Kubeconfig | Server address |
|---|---|---|
| Your machine | `<UserConfigDir>/cka-sim/kubeconfig` (e.g. `~/.config/cka-sim/kubeconfig`), context `kind-cka-sim` | `https://127.0.0.1:<port>`, a port Docker maps to the API server |
| Inside the control plane | `/etc/kubernetes/admin.conf` | `https://cka-sim-control-plane:6443`, the name on the Docker network |

cka-sim writes its own file and never touches `~/.kube/config`, so it can't break a user's other clusters.

Run:
```sh
kubectl --kubeconfig ~/.config/cka-sim/kubeconfig get nodes                 # host, if you have kubectl
grep server ~/.config/cka-sim/kubeconfig                                     # host
docker exec cka-sim-control-plane grep server /etc/kubernetes/admin.conf     # host
docker port cka-sim-control-plane                                            # host: the mapped port
```

**Answer**
1. Why do the two `server:` addresses differ? (Port mapping to your host vs the Docker network.)
2. Which one would work from another container on the `kind` network? Which one from your host?

---

### Step 6 · The Go program so far

The code follows the [conventions](BUILD-FROM-ZERO.md#conventions): it should explain itself, so a
comment means there is something you can't see from the code alone. When you find a comment, ask
what it tells you that the code doesn't.

| File | Concepts to look for |
|---|---|
| `cmd/cli/main.go`, `cmd/cli/cli.go` | a `switch` on the command; `signal.NotifyContext` turns Ctrl-C into cancellation; `runUp` races kind's uncancellable `Create` against `ctx.Done()`; the version from `-ldflags` or `debug.ReadBuildInfo` |
| `internal/doctor/doctor.go` | a struct of functions (`System`) as the seam for tests; string-typed `Status` constants; decoding only the fields we need from `docker info` JSON |
| `internal/cluster/cluster.go` | kind's public API (`NewProvider`, `Create` options); the config as plain YAML; `os.UserConfigDir` for a per-OS config folder |
| `.github/workflows/ci.yml` | a test matrix over three operating systems; cross-compiling six targets with `CGO_ENABLED=0` |

Run:
```sh
cd apps/cka-sim && go vet ./... && go test ./...                  # host
go mod why -m github.com/spf13/cobra                          # host: why an indirect dependency is there
```

**Answer**
1. `runUp` returns on Ctrl-C while kind keeps creating containers in a goroutine. What does the user have to do next, and how do they find out?
2. Why does `cluster_test.go` parse the YAML into kind's own types instead of checking the text with `strings.Contains`?
3. Every dependency in `go.mod` except `sigs.k8s.io/yaml` is `// indirect`. What does that mean?

---

### Step 7 · Anatomy of a task *(v1 notes, rewritten by D4–D5)*

- [ ] Read on `archive/v1`: the five files of `tasks/troubleshooting/tr-service/`, then `tasks/lib.sh`.

```
task.md       frontmatter (id, domain, weight, …) + the question in exam wording
setup.sh      puts the cluster into its broken or starting state; must be re-runnable
check.sh      prints PASS/FAIL lines
solution.sh   the reference fix, with the diagnosis in comments
explain.md    the concept, shown after the exam
```

The grader runs somewhere the candidate can't reach. In v1, `lib.sh` gave scripts:
- `k`: kubectl with the task's context
- `on <node> <cmd>`: `docker exec` into a node as root
- `check <points> <description> <command>`: runs the command and prints `PASS` or `FAIL`
- `wait_for <seconds> <command>`: retries until the command succeeds
- `fresh_ns <ns>`: force-deletes leftovers, then recreates the namespace
- `remember` / `recall`: state passed from setup to check, kept where the candidate can't see it

v1 ran these with bash **on the host**. The rebuild must run them in a container instead, because
Windows has no bash (see D5).

**Answer**
1. How did `tr-service`'s check prove you didn't "fix" it by editing the Deployment?
   (`remember generation` in setup; compared in check.)
2. Why does `fresh_ns` force-delete Pods before deleting the namespace?
   (A Pod on a node with a dead kubelet never confirms termination.)

---

### Step 8 · Grading: a tiny text protocol *(v1 notes, rewritten by D6)*

- [ ] Read on `archive/v1`: `internal/grader/grader.go` and `grader_test.go`.

`check.sh` prints lines like `PASS 3 Service web has endpoints`. The parser keeps only lines that
match `PASS|FAIL <int> <text>` and ignores the rest, so scripts can print diagnostics freely. This
is the idea behind TAP (Test Anything Protocol): agree on a line format, and any language can be a test.

- **Partial credit:** score = Σ(task weight × earned/total) ÷ Σ weights.
- **Gating:** constraint checks only pass once the main goal does, otherwise doing nothing earns points.

**Answer**
1. Why was `pipefail` removed from the runner? (Hint: `grep -q` exits early → SIGPIPE upstream.)
2. What happens if `check.sh` hangs? (v1 used a 3-minute `context.WithTimeout`.)

---

### Step 9 · The selftest: proving every task is fair *(v1 notes, rewritten by D7)*

The rule for every task: **after setup it scores 0; after `solution.sh` it scores full marks.**

The bugs it caught in v1:
- **Free points:** constraint checks passed on untouched tasks, which led to gating.
- **Incomplete etcd restore:** restoring etcd under a running API server left it serving stale
  objects from its watch cache. The fix: stop the API server, restore with
  `--bump-revision --mark-compacted`, then start it again.
- **Interfering tasks:** a Pod landed on the node another task breaks, which jammed the
  namespace. The fix is a `nodeSelector`, plus `fresh_ns` force-deleting Pods.

**Answer**
1. Why must an etcd restore task run last? What would a restore do to tasks set up after the snapshot?
2. Write one invariant like this for a project of your own.

---

### Step 10 · Making a node look like an exam host *(v1 notes, rewritten by D9)*

- [ ] Read on `archive/v1`: `images/node/Dockerfile` and `images/node/exam-profile.sh`.

The v1 image started `FROM kindest/node` and added:
- **sshd**, a `candidate` user with passwordless sudo, and password login disabled
- **man pages:** the slim base deletes them through `/etc/dpkg/dpkg.cfg.d/docker`, so the image
  removes that file, reinstalls the packages and runs `mandb`
- **yq, etcdctl and etcdutl**, with etcd pinned to the version kubeadm uses
- **the `k` alias and completion**, loaded from `/etc/bash.bashrc`

The rebuild publishes this image to GHCR, so users pull it instead of building it.

**Answer**
1. Why does `ssh node "k get nodes"` fail when an interactive `ssh node` followed by `k get nodes` works?
   (Aliases only load in interactive shells.)
2. Why pin the etcd version instead of using the latest?

---

### Step 11 · Container networking: how `ssh <node>` works *(v1 notes, rewritten by D10)*

- [ ] Read on `archive/v1`: `startBase()` and `provisionNodes()` in `internal/env/env.go`, plus `images/base/ssh_config`.

Every container joins the Docker network `kind`. On a user-defined Docker network, Docker runs an
**embedded DNS server (127.0.0.11)** that resolves container names to their IPs. That's the whole
trick: the container name is the hostname.

The ssh trust was wired like this:
1. Generate an ed25519 key **inside base**.
2. Write that public key to every node's `/home/candidate/.ssh/authorized_keys`.
3. `ssh_config` sets `User candidate` and skips host-key checks, because nodes are rebuilt all the time.

Run (works today, before base exists):
```sh
docker network inspect kind --format '{{range .Containers}}{{.Name}} {{.IPv4Address}}{{"\n"}}{{end}}'   # host
docker exec cka-sim-worker getent hosts cka-sim-control-plane                                        # host
```

**Answer**
1. Would this work on Docker's *default* `bridge` network? (No: it has no embedded DNS.)
2. Where does base's private key live, and why is it never copied out of base?
3. What would you change to make nested ssh (node → node) impossible, as on the exam?

---

### Step 12 · The CNI and NetworkPolicy *(v1 notes, rewritten by D8)*

kind ships its own CNI, **kindnet**. On the WSL2 kernel in v1, kindnet's NetworkPolicy enforcement
needed `CONFIG_NFT_QUEUE`, which wasn't there, so policies were *silently allowed*. A NetworkPolicy
task that grades on traffic was then impossible. **Calico** enforces policy with iptables and worked.

Two settings made Calico fit kind:
- `disableDefaultCNI: true` and `podSubnet: 192.168.0.0/16`, which is Calico's default pool.
- A strategic-merge patch that deletes Calico's `/sys/kernel/security` mount, because WSL2 has no
  securityfs. Apply the manifest only once, since re-applying brings that mount back.

D8 starts by re-testing kindnet on kind v0.33, because the problem may be gone.

Run (works today):
```sh
docker exec cka-sim-control-plane kubectl -n kube-system get pods -o wide | grep kindnet   # host
```

**Answer**
1. Which component gives a Pod its IP: the kubelet, containerd, or the CNI plugin?
2. Why are nodes `NotReady` until the CNI is installed?
3. Why does a strategic-merge patch with `$patch: delete` work on lists that a JSON merge patch would replace?

---

### Step 13 · The panel *(v1 notes, rewritten by D14)*

- [ ] Read on `archive/v1`: `web/index.html` (the `<script>` part) and `internal/server/server.go`.

- **No framework:** `render()` rebuilt the main area from the exam state, and every action called
  the API and re-rendered with its response.
- **Timer:** the server sends `now`; the page stores the offset to its own clock and counts
  locally, so a wrong browser clock can't change your deadline.
- **Persistence:** all state was in `exam.json`, so the browser could close at any time.
- **Study mode is enforced on the server**, not just hidden in the UI.

**Answer**
1. In an exam, where would a cheater look for solutions, and why don't they find them?
   (In v1's `view()`, solutions were only filled in when the exam had ended or in study mode.)

---

### Environment quirks

| Symptom | Root cause | Status |
|---|---|---|
| kubelet refused to start | cgroup v1 on WSL2; Kubernetes ≥ 1.35 refuses it by default | handled in D2 (step 4) |
| NetworkPolicy allowed everything | kindnet needed `CONFIG_NFT_QUEUE` on WSL2 | re-check in D8 |
| calico-node never started | no securityfs on WSL2 | D8, if Calico is needed |
| image pulls failed | Docker Desktop's credential helper broken in `~/.docker/config.json` | v1 used a private empty `DOCKER_CONFIG`; not seen in the rebuild yet |

---

### Checkpoint: explain these out loud

- [ ] What doctor checks, and why CPU and memory come from Docker and not from the host.
- [ ] What happens between `cka-sim up` and three Ready nodes, step by step.
- [ ] What a static Pod is, and who restarts it when you edit its manifest.
- [ ] What cgroups are, and what `failCgroupV1: false` does and doesn't change.
- [ ] Why your host and the control plane use different `server:` addresses for the same cluster.

Later, as the rebuild gets there:
- [ ] How a `check.sh` line turns into a score. (D6)
- [ ] Why NetworkPolicy depends on the CNI. (D8)
- [ ] How `ssh cka-sim-worker` reaches the right container, starting from DNS. (D10)
- [ ] Why an etcd restore needs the API server stopped. (D16)

---

## Part 2 — Deploy it

The goal: use cka-sim from any browser, not just your laptop. This needs the panel (D14) and the
exam command (D15). Do the phases in order; each one works on its own.

> ⚠ **Security model, read this first.** kind nodes are **privileged** containers. A shell on
> one is effectively **root on the machine that runs Docker**. That's fine when the machine
> is yours and only you use it. It is **not** fine for strangers sharing one Docker host.
> Phases 1–4 are single-user for this reason; Phase 5 is what multi-user takes.

### Phase 1 · One cloud VM, by hand

- [ ] Create a Linux VM: Ubuntu 24.04, at least **4 vCPU / 8 GB RAM**, with 16 GB to be
      comfortable; about 40 GB disk. The cost adds up only while it runs.
- [ ] Install Docker and cka-sim (D17), then `cka-sim doctor && cka-sim up`.

What's different from WSL2: a cloud VM normally runs **cgroup v2** and has securityfs, so the WSL2
workarounds don't kick in. Doctor should show every line `[ok]`.

**Learn:** VM sizing, how kind behaves on a "real" Linux host.

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
      and it includes a terminal, so without authentication anyone who finds the URL gets a shell.
- [ ] A firewall: allow only 22 and 443.

**Learn:** port forwarding, reverse proxies, TLS certificates, why services bind to localhost.

### Phase 3 · A terminal in the browser

The panel's terminal (D14) is xterm.js in the page, a websocket endpoint in the Go server, and a
pty running the same command as `cka-sim shell`. Once Phase 2 puts the panel behind your proxy and
authentication, the browser is all you need.
- [ ] Make sure the proxy forwards websockets (`Upgrade` headers). Caddy does this by default.

**Learn:** websockets, pseudo-terminals (pty), how browser terminals work.

### Phase 4 · Automate the VM

- [ ] **cloud-init** (or a shell script) that installs Docker and cka-sim and runs `cka-sim up` on first boot.
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
| **Cost:** several GB of RAM per session | session limits, pre-warmed pools, shorter exams |
| **Accounts:** who is who | real authentication; exam state per user instead of one session file |
| **Running the platform on Kubernetes** | an **operator**: a CRD `ExamSession`, and a controller that creates and cleans up each session's VM |

**Learn:** multi-tenancy, sandboxing, controllers and CRDs (the deepest Kubernetes topic).

---

### Deployment checkpoint

- [ ] I can reach the panel from another device over HTTPS, with a login.
- [ ] I can work a task entirely in the browser.
- [ ] I can destroy and recreate the whole VM with one command.
- [ ] I can explain why Phase 5 needs a VM per user, and not a container per user.
