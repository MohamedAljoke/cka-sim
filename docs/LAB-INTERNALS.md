# How the lab works inside

Most of cka-sim is an ordinary API, page and use cases. This file covers the three parts that aren't ordinary:

1. [A whole Kubernetes cluster inside one container](#1-docker-in-docker)
2. [How the platform reaches it](#5-platform--agent-connect-over-http2)
3. [How a terminal in the browser becomes a shell on an exam-like host](#6-the-terminal-a-pty-and-docker-exec)

**Contents:** [Big picture](#the-big-picture) · [1. Docker in Docker](#1-docker-in-docker) ·
[2. kind](#2-kind-a-cluster-made-of-containers) · [3. The base host and ssh](#3-the-base-host-and-ssh) ·
[4. Cluster state](#4-the-clusters-state-lives-in-one-process) · [5. Platform ⇄ agent](#5-platform--agent-connect-over-http2) ·
[6. The terminal](#6-the-terminal-a-pty-and-docker-exec) · [7. Who can reach what](#7-who-can-reach-what) ·
[8. Try it](#8-try-it-yourself)

## The big picture

![The kind cluster and the base host run on a private Docker inside the lab container](img/lab-container.svg)

| Part | Where | What it does |
| --- | --- | --- |
| Browser | xterm.js | Draws the terminal and talks to the platform over the WebSocket `/ws/terminal` |
| Platform | `:7070` | Serves the API and translates the WebSocket into `LabService` calls, with a bearer token |
| Agent | `:7171`, in the lab container | Builds the cluster and the base host, opens shells on base, runs task scripts on nodes |
| Inner dockerd | in the lab container | Runs the kind nodes and the base host. It listens on a Unix socket only |
| Base host | `cka-sim-base` | Where every terminal lands, as `candidate`, with no kubectl. `ssh <node>` from there |
| `lab-docker` volume | `/var/lib/docker` | Keeps the pulled and built images between runs |

![Three levels of containers](img/nesting.svg)

> [!NOTE]
> The platform never talks to Docker. It only talks to the agent, so a lab could later run on a VM or GKE and
> nothing in the platform would change.

## 1. Docker in Docker

The lab image is `docker:28-dind` plus the agent (`infra/lab/lab.Dockerfile`). `lab-start.sh` starts dockerd,
waits until `docker info` answers, and then execs the agent.

| Problem | Fix |
| --- | --- |
| dockerd and kind's nodes create cgroups, mount filesystems and change networking, which a normal container can't do | `privileged: true` |
| `kindest/node` is about 1 GB | The `lab-docker` volume keeps `/var/lib/docker`, so only the first cluster downloads it |
| Three kubelets run out of the kernel's default inotify watches | `lab-start.sh` raises `fs.inotify.max_user_watches` and `max_user_instances` |
| A restarted container keeps `/var/run`, and dockerd then waits forever for a containerd that is gone | `lab-start.sh` deletes the stale pid files first |
| Docker must not be reachable from outside | dockerd listens on a Unix socket only; only the agent uses it |

```sh
docker exec cka-sim-lab-1 docker ps    # base and the nodes, one level down
```

## 2. kind: a cluster made of containers

kind ("Kubernetes in Docker") makes each node a container from a node image, which has systemd, containerd and a
kubelet. It runs `kubeadm init` on the control plane and `kubeadm join` on the workers. The result is a real
cluster, with a real etcd, static pods and kubelets you can break and fix, which is exactly what CKA tasks need.

**kind is a Go library here, not the `kind` CLI** (`apps/lab/internal/infra/adapters/kind/kind.go`):

```go
kindcluster.NewProvider(ProviderWithDocker(), ProviderWithLogger(stepLogger{progress}))
```

- `Create` gets our node image, the cluster config, a kubeconfig path (`~/.cache/cka-sim/<name>.kubeconfig`, so
  `~/.kube/config` stays untouched) and `WaitForReady(5m)`.
- kind still drives Docker through the `docker` CLI, so it follows the current Docker context, inside the lab
  container or on your machine alike.
- `kind.yaml` (1 control plane, 2 workers), the cgroup v1 patch and both Dockerfiles are embedded with
  `//go:embed`, so the binary carries everything it builds.

**An exam-like node image.** Before kind starts, the agent builds `cka-sim/node:v1.37.0` from `node.Dockerfile`
with `docker build -`, inside the lab's own Docker. On top of `kindest/node` it adds:

- sshd, enabled through systemd (kind's entrypoint has to stay PID 1, so sshd can't be the `CMD`), with password
  logins off
- a `candidate` user with sudo and no password
- vim, bash completion, `alias k=kubectl` and kubectl completion for `k`

A unit test checks that the tag matches the `FROM` line, so the two can't drift apart.

**cgroups.** Kubernetes 1.35+ kubelets refuse cgroup v1 unless the config sets `failCgroupV1: false`. The agent
asks `docker info` which version the host uses and adds the patch only for v1. WSL2 and modern Linux use v2.

**Live progress, with no log file.** kind reports its phases by calling the logger it was given:

```
kind logs " ✓ Preparing nodes 📦"  →  stepLogger strips it to "Preparing nodes"  →  progress(step)  →  Status.Step
```

`progress` is `cluster.Progress`, which overwrites one string under a mutex. The agent adds its own steps the
same way: "Building the node image", "Waiting for every node to be Ready", "Starting the base host". No
`docker logs` and no polling: just function calls inside the agent process.

**kind only waits for the control plane.** After `Create` returns, `waitForNodes` runs `kubectl` in the
control-plane container until every node is Ready.

**Every start is fresh.** At startup the agent deletes any cluster with its name. This process didn't watch a
cluster from before a restart, so its state is unknown.

## 3. The base host and ssh

The CKA exam starts you on a machine with no cluster access, and each task says which node to `ssh` to first.
cka-sim copies that with a fourth container next to the nodes (`apps/lab/internal/infra/adapters/kind/base.go`).

![The base host reaches every node with ssh](img/base-host.svg)

After every node is Ready, `startBase`:

1. builds `cka-sim/base` from `base.Dockerfile`: Ubuntu with an ssh client, vim and curl, a `candidate` user
   with no sudo, and **no kubectl**
2. runs it as `<cluster>-base` with `--network kind`, so node names like `cka-sim-worker` resolve, and `--init`,
   so the processes ended shells leave behind are reaped
3. makes an ed25519 key pair for `candidate` on base
4. copies to every node, with `docker exec -i … install`:

| File on each node | Owner | Why |
| --- | --- | --- |
| `/home/candidate/.ssh/authorized_keys` | candidate | base's public key, so `ssh <node>` needs no password |
| `/home/candidate/.kube/config` | candidate | kubectl works right after `ssh`, on any node |
| `/root/.kube/config` | root | `setup.sh` and `check.sh` run as root on the task's node, and only the control plane came with one |

Base's ssh config sets `User candidate` and turns off host key checking. Every new cluster recreates the
nodes with new host keys, and a "host key changed" warning would only get in the way.

**Where things run:**

| What | Where | As |
| --- | --- | --- |
| The page's terminal and `cka-sim shell` | base | candidate |
| `ssh <node>` from there | that node | candidate, `sudo` for root |
| `RunScript` (setup and check) | the task's node, never base | root |

`RunScript` only accepts a host from `Nodes`, so asking for base, or any other container, is INVALID_ARGUMENT
(`ErrUnknownHost`). `Delete` removes base first, then the cluster.

## 4. The cluster's state lives in one process

![The cluster's state machine](img/cluster-states.svg)

`apps/lab/internal/domain/cluster/cluster.go` is a small state machine guarded by a mutex. Ready means the
nodes and base are both up: `Create` returns only after `startBase`.

| Rule | Why |
| --- | --- |
| Repeats are safe: starting twice returns the current status | Only the caller that gets `true` from `BeginCreate` launches the build |
| Conflicts are refused: deleting while Creating returns `ErrBusy` | The API answers FAILED_PRECONDITION |
| Late callbacks are ignored | A progress line after a failure changes nothing |
| It lives in memory | Fine for one lab per agent |

The build runs in a goroutine with `context.WithoutCancel` and a 20-minute limit, so it outlives the HTTP call
that started it.

## 5. Platform ⇄ agent: Connect over HTTP/2

The contract is `apps/go-shared/contract/lab/v1/lab.proto`. Each app knows only the contract, never the other
app's code, and `make proto` regenerates it.

| Call | Kind | Transport |
| --- | --- | --- |
| `GetCluster`, `CreateCluster`, `DeleteCluster` | unary | Plain HTTP POSTs. They work over HTTP/1.1, even as JSON |
| `RunScript` | unary | The same; bash as root on one of the nodes |
| `OpenShell` | bidi stream | One call stays open as long as the terminal. Needs HTTP/2 |

A unary call from the shell:

```sh
curl -H "Authorization: Bearer $T" -H 'Content-Type: application/json' -d '{}' \
  http://127.0.0.1:7171/lab.v1.LabService/GetCluster
```

### Why the stream needs HTTP/2

HTTP/1.1 is request, then response: a client can't keep sending a body while it reads the answer. That's why
WebSocket exists. HTTP/2 is different: each request is a stream of frames inside one TCP connection, and both
bodies can stay open at the same time.

![OpenShell as one HTTP/2 stream](img/http2-stream.svg)

The last message carries the status, which is how FAILED_PRECONDITION reaches the caller. Connect reports a clean
end as an error that matches `errors.Is(err, io.EOF)`.

### h2c: HTTP/2 without TLS

Both sides have to opt in:

| Side | How |
| --- | --- |
| Agent | `protocols.SetUnencryptedHTTP2(true)` in `cmd/agent/main.go` |
| Platform | `http.Transport{Protocols: …SetUnencryptedHTTP2(true)}` |

> [!WARNING]
> A streaming client must not set `http.Client.Timeout`, because it would cut every shell off after that long.

**The browser hop stays WebSocket**, because browser JavaScript can't stream a request body over HTTP/2. The
platform translates:

| WebSocket | `OpenShell` |
| --- | --- |
| binary frame from the browser | `keys` |
| binary frame to the browser | `output` |
| text frame `{"type":"resize",cols,rows}` | `size` |

## 6. The terminal: a pty and `docker exec`

### What a pty is

A pseudo-terminal is a pair of file descriptors from the kernel:

- **slave:** looks like a real terminal. A program gets it as stdin, stdout and stderr and believes it is
  interactive.
- **master:** what our code holds. Bytes written there arrive as if typed, and whatever the program prints comes
  out there.

`docker exec -it` refuses to run without a terminal ("the input device is not a TTY"), and bash without one has
no prompt or line editing. So the agent starts the CLI under a pty with `github.com/creack/pty`
(`apps/lab/internal/infra/adapters/dockerexec/dockerexec.go`):

```sh
docker exec -it -e CKA_SIM_TERMINAL=<tag> -u candidate -w /home/candidate <cluster>-base bash -l
```

![Two terminals between the browser and bash](img/pty.svg)

| Terminal | Who makes it | What it does |
| --- | --- | --- |
| Our pty | the agent, with creack/pty | The docker CLI puts it in raw mode, so it passes bytes straight through |
| Base's terminal | dockerd, because of `-t` | bash is attached to it, and bash does the echo and line editing |

After `ssh <node>`, there is a third: ssh asks sshd on the node for one, relays bytes to it over the kind
network, and forwards window size changes, so `vim` on a node redraws just like on base.

### One keystroke, end to end

![One keystroke from xterm.js to bash on the base host and back](img/terminal.svg)

```
key → xterm.js → WebSocket → platform → OpenShell {keys} → agent → pty master → pty slave
    → docker CLI → dockerd → base's terminal → bash on base   (→ ssh → sshd → bash on the node)
bash echoes "l" → the same way back → {output} → WebSocket → xterm.js draws "l"
```

> [!TIP]
> The browser never draws what you type. Every character you see, even a space, is bash's echo coming back.
> That's why passwords stay invisible, and why a slow network makes typing lag.

### Resize

Programs ask *their* terminal how big it is, so the browser panel's size is copied all the way down:

1. The fit add-on measures cols × rows and the page sends a resize message.
2. The agent gets `{size}` and calls `pty.Setsize`.
3. The kernel sends SIGWINCH to the docker CLI, which asks dockerd to resize base's terminal.
4. The kernel sends SIGWINCH to whatever runs there. bash, vim and top redraw; an ssh client passes the new size on
   to the node.

**The shell starts only at the first size** (`pty.StartWithSize`). A resize sent while the shell is still
starting can get lost, and the first prompt would be laid out for the wrong width. That's why the contract says
the first message is the size.

### Ending a shell

Killing the `docker exec` client does **not** kill what it started: bash keeps running on base. So every shell
gets a random tag in its environment. `Close` kills the CLI, then runs a script on base that finds every process
whose `/proc/<pid>/environ` holds that tag and sends it HUP, then KILL. An `ssh` the shell started carries the tag
too, and when it dies, sshd ends the session on the node. The idea comes from the first version's `dockershell.go` (`archive/v2`).

On the agent (`apps/lab/internal/api/shell.go`), each direction has its own goroutine, because Connect allows one
`Send` and one `Receive` at a time:

![One goroutine per direction](img/shell-pumps.svg)

## 7. Who can reach what

| From | Can reach | Can't reach |
| --- | --- | --- |
| The terminal, on base | every node over ssh, as candidate; the agent's port over the network | the cluster directly (no kubectl, no kubeconfig); any RPC, because it has no token |
| candidate on a node | root, with `sudo` and no password | |
| Root on a privileged node | the lab container, by breaking out, and the lab container is privileged too | |
| Any website you visit | nothing, because the WebSocket checks the `Origin` | |

> [!IMPORTANT]
> **The shell user is not the boundary.** CKA tasks need root on the nodes, so `sudo` is one command away, and
> from there a privileged node can reach the lab container. **One lab container per learner is the boundary**,
> plus whatever isolates that container.

The origin check matters because a GET is a "safe" method, so `CrossOriginProtection` lets the WebSocket upgrade
through. Without the check, any site could open a shell through your platform.

## 8. Try it yourself

**The terminal without the page**, once a lab is Ready:

```sh
cka-sim shell          # installed: candidate on base, like the page
ssh cka-sim-worker     # then kubectl works there
```

**The lab in its container** (agent on 7171):

```sh
make lab
docker exec cka-sim-lab-1 docker ps                                            # base and the nodes
docker exec cka-sim-lab-1 docker exec cka-sim-control-plane kubectl get nodes  # the cluster
```

**The agent straight on your machine's Docker**, beside the container:

```sh
cd apps/lab && go build -o /tmp/agent ./cmd/agent
CKA_SIM_AGENT_ADDR=127.0.0.1:7272 CKA_SIM_CLUSTER=cka-sim-local \
  CKA_SIM_AGENT_TOKEN=local-only-agent-token-not-a-secret /tmp/agent

T=local-only-agent-token-not-a-secret
curl -s -H "Authorization: Bearer $T" -H 'Content-Type: application/json' -d '{}' \
  http://127.0.0.1:7272/lab.v1.LabService/CreateCluster
docker ps    # cka-sim-local-base, cka-sim-local-control-plane and its workers
```

**A real cluster through the agent** (takes a few minutes). The test opens a shell, checks it is `candidate` with
no kubectl, then runs `kubectl get nodes` over `ssh` on a worker:

```sh
cd apps/lab && go test ./e2e -create
```
