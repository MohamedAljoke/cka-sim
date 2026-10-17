# How the lab works inside

Most of cka-sim is the usual stack: an API, a page and a few use cases. This file covers what is new in this
project:
- how a whole Kubernetes cluster runs inside one container
- how the platform reaches it
- how a terminal in the browser becomes a root shell on a node

## The big picture

```
your machine
├─ browser: xterm.js
│     │ WebSocket /ws/terminal                   (browser ⇄ platform)
├─ platform :7070
│     │ Connect over HTTP/2: LabService          (platform ⇄ agent, bearer token)
└─ lab container (privileged, docker:dind)       infra/local/compose.yml
   ├─ cka-sim-agent :7171                        apps/lab/cmd/agent
   │     ├─ kind as a Go library ─┐
   │     └─ docker exec under pty ┤
   ├─ dockerd  ◀──────────────────┘ /var/run/docker.sock
   │  ├─ cka-sim-control-plane     (a "node": systemd, containerd, kubelet, etcd, apiserver)
   │  ├─ cka-sim-worker
   │  └─ cka-sim-worker2
   └─ volume lab-docker → /var/lib/docker (keeps the pulled node image)
```

There are three levels of containers. Your Docker runs the lab container. The lab container runs its own
dockerd, which runs the kind nodes. Each node runs containerd, which runs the cluster's pods. The platform never
talks to Docker. It only talks to the agent, so a lab could later run anywhere, such as a VM or GKE, and
nothing about the platform would change.

## 1. Docker in Docker

The lab image is `docker:28-dind` plus the agent (`apps/lab/deploy/lab.Dockerfile`). `lab-start.sh` starts
dockerd, waits until `docker info` answers, and then execs the agent.

- **Why `privileged: true`:** dockerd and kind's nodes need to create cgroups, mount filesystems and change
  networking. A normal container isn't allowed to do any of that.
- **Why the volume:** `kindest/node` is about 1 GB. Keeping `/var/lib/docker` in the `lab-docker` volume
  means only the first cluster downloads it.
- **inotify limits:** every node runs a kubelet, and three of them run out of the kernel's default inotify
  watches. `lab-start.sh` raises `fs.inotify.max_user_watches` and `max_user_instances`.
- **Stale pid files:** a restarted container keeps `/var/run`, and dockerd then waits forever for a containerd
  that no longer exists. `lab-start.sh` deletes those pid files first.
- **dockerd listens on a unix socket only.** Nothing outside the container can reach Docker, only the agent.

Try it: `docker exec cka-sim-lab-1 docker ps` lists the nodes inside the lab container.

## 2. kind: a Kubernetes cluster made of containers

kind ("Kubernetes in Docker") makes each node a container from `kindest/node`. The image has systemd,
containerd and kubelet. kind runs `kubeadm init` on the control-plane container and `kubeadm join` on the
workers. The result is a real cluster: real etcd, real static pods, and real kubelets you can break and fix,
which is exactly what CKA tasks need.

**We use kind as a Go library, not the `kind` CLI** (`apps/lab/internal/infra/adapters/kind/kind.go`):

- `kindcluster.NewProvider(ProviderWithDocker(), ProviderWithLogger(stepLogger{progress}))`
- `Create` is given the cluster config, a kubeconfig path (`~/.cache/cka-sim/<name>.kubeconfig`, so
  `~/.kube/config` stays untouched) and `WaitForReady(5m)`.
- kind still drives Docker by running the `docker` CLI. So it follows the current Docker context, inside the
  lab container or on your machine alike.

**Config and patches are compiled in.** `kind.yaml` (1 control plane + 2 workers) and
`kind-cgroupv1-patch.yaml` are embedded with `//go:embed`, so the binary carries them.

**cgroups.** Kubernetes 1.35+ kubelets refuse cgroup v1 unless the config sets `failCgroupV1: false`. The
provisioner asks `docker info` which version the host uses and adds the patch only for v1. WSL2 and modern
Linux use v2.

**Live progress without a log file.** kind reports its phases by calling the logger we gave it, for example
`" ✓ Preparing nodes 📦"`. `stepLogger` (`logger.go`) strips that down to `Preparing nodes` and calls
`progress(step)`, which is `cluster.Progress`. That overwrites one `Status.Step` string under a mutex. No
`docker logs` and no polling of kind: it's plain function calls inside the agent process.

**kind only waits for the control plane.** After `Create` returns, `waitForNodes` uses `kubectl` inside the
control-plane container to check that every node is Ready.

**Start fresh.** At startup the agent deletes any cluster with its name. A cluster from before a restart
wasn't watched by this process, so its state is unknown.

## 3. The cluster's state lives in one process

`apps/lab/internal/domain/cluster/cluster.go` is a small state machine guarded by a mutex:

```
Absent ──BeginCreate──▶ Creating ──Created──▶ Ready ──BeginDelete──▶ Deleting ──Deleted──▶ Absent
                           │ CreateFailed                               │ DeleteFailed
                           ▼                                            ▼
                         Failed  (BeginCreate and BeginDelete both work from here)
```

- **Repeats are safe.** Starting twice returns the current status, and only the caller that gets `true`
  from `BeginCreate` launches the build.
- **Conflicts are refused.** Deleting while Creating returns `ErrBusy`, which the API turns into
  FAILED_PRECONDITION.
- **Late callbacks are ignored.** A progress line arriving after a failure changes nothing.
- **It lives in memory.** That's fine for one lab per agent. The build runs in a goroutine with
  `context.WithoutCancel`, so it outlives the HTTP call that started it, with a 20-minute limit.

## 4. Platform ⇄ agent: Connect, and why HTTP/2

The contract is `apps/go-shared/contract/lab/v1/lab.proto`. Each app only knows the contract, never the
other app's code, and `make proto` regenerates it.

- **Unary calls** (`GetCluster`, `CreateCluster`, `DeleteCluster`) are ordinary HTTP POSTs. They work over
  HTTP/1.1, and even as JSON:
  `curl -H "Authorization: Bearer $T" -H 'Content-Type: application/json' -d '{}' http://127.0.0.1:7171/lab.v1.LabService/GetCluster`
- **`OpenShell` is a bidi stream:** one call stays open as long as the terminal, with many messages each way.

**How one HTTP request can stream both ways.** HTTP/1.1 is request, then response, so a client can't keep
sending a body while it reads the answer. That's why WebSocket exists: it upgrades the connection to its own
protocol. HTTP/2 is different. Each request is a stream of frames inside one TCP connection, and both bodies
can stay open together:

```
stream 1: POST /lab.v1.LabService/OpenShell
  → HEADERS  (Authorization)          ← HEADERS 200
  → DATA [size 120x40]                ← DATA [output "root@…:~# "]
  → DATA [keys "l"]                   ← DATA [output "l"]
  …                                   ← end-of-stream message with the status
```

Connect frames each message as `[1 flag byte][4-byte length][protobuf]`, so `Receive()` returns one whole
message. The last envelope carries the status, which is how a FAILED_PRECONDITION reaches the caller.
Connect reports a clean end as an error that matches `errors.Is(err, io.EOF)`.

Running HTTP/2 without TLS is called **h2c**. Both sides have to opt in:

- the agent with `protocols.SetUnencryptedHTTP2(true)` (`cmd/agent/main.go`)
- the caller with an `http.Transport{Protocols: …SetUnencryptedHTTP2(true)}`

A streaming client must not have an `http.Client.Timeout`, because the timeout would cut every shell off
after that long.

**The browser hop stays WebSocket**, because browser JavaScript can't stream a request body over HTTP/2.
The platform translates:

- WebSocket binary frames ⇄ `keys` and `output`
- `{"type":"resize",cols,rows}` text frames ⇄ `size`

## 5. The terminal: a pty and `docker exec`

### What a pty is

A pseudo-terminal is a pair of file descriptors from the kernel:

- **slave:** looks like a real terminal. A program gets it as stdin, stdout and stderr and believes it is
  interactive.
- **master:** what our code holds. Bytes written there arrive as if typed, and whatever the program prints
  comes out there.

`docker exec -it` refuses to run without a terminal ("the input device is not a TTY"). Bash without one has no
prompt or line editing, and Ctrl-C, `vim` and `top` don't work. So the agent starts the CLI under a pty with
`github.com/creack/pty` (`apps/lab/internal/infra/adapters/dockerexec/dockerexec.go`):

```
docker exec -it -e CKA_SIM_TERMINAL=<tag> -e KUBECONFIG=/etc/kubernetes/admin.conf -w /root \
  <cluster>-control-plane bash -l
```

### The whole chain for one keystroke

```
key → xterm.js onData → WebSocket → platform → OpenShell {keys} → agent feed() → pty master
    → pty slave → docker CLI → dockerd → terminal inside the node → bash (readline)
bash echoes "l" → … the same way back … → agent Read → {output} → WebSocket → xterm.js draws "l"
```

Two terminals are involved:

- **Our pty:** the docker CLI puts it in raw mode, so it passes bytes straight through.
- **The node's terminal:** created by `-t`. Bash is attached to it, and bash does the echo and line editing.

**The browser never draws what you type.** Every character you see, even a space, is bash's echo coming back.
That's why `sudo` passwords stay invisible, and why a slow network makes typing lag.

### Resize

Programs in the node ask *their* terminal how big it is, so the browser panel's size has to be copied all the
way down:

```
fit add-on measures cols×rows → resize message → {size} → pty.Setsize
  → kernel sends SIGWINCH to the docker CLI → CLI asks dockerd to resize the node's terminal
  → kernel sends SIGWINCH to bash/vim/top → they redraw
```

**The shell starts only at the first size** (`pty.StartWithSize`). A resize sent while the shell is still
starting can get lost, and the first prompt would be laid out for the wrong width. So the contract says the
first message is the size.

### Ending a shell, and the leftover-bash problem

Killing the `docker exec` client does **not** kill what it started: bash keeps running in the node. So every
shell gets a random tag in its environment. `Close` kills the CLI, then runs a script in the node that finds
every process whose `/proc/<pid>/environ` contains that tag and sends it HUP, then KILL. This approach comes
from `main`'s `dockershell.go`.

On the agent (`apps/lab/internal/api/shell.go`), each direction has its own goroutine, because Connect allows
one Send and one Receive at a time:

```
caller ─Receive─▶ feed()       ─Write/Resize─▶ pty      (ends: caller hung up → Close the shell)
caller ◀──Send── handler loop  ◀────Read────── pty      (ends: bash exited → handler returns)
```

Either side ending ends the other one:

- Closing the shell makes the handler's `Read` fail.
- The handler returning ends the stream, so `feed`'s `Receive` fails.

`Close` therefore runs twice, and `sync.Once` makes the second call a no-op.

## 6. Who can reach what

- A shell in a node can see the node's processes, not the agent's.
- It can reach the agent's port over the network, but every RPC needs the bearer token, which lives only in
  the agent's environment.
- **Root on a privileged node can break out** into the lab container, and the lab container is privileged too.
  The base host design (step 2.3) has the same reach through `sudo`, which CKA tasks need. So the shell user
  is not the boundary. **One lab container per learner is**, plus whatever isolates that container.
- The browser hop needs an origin check: a GET is a "safe" method, so `CrossOriginProtection` lets it through.
  Without the check, any site you visit could open a root shell through your platform.

## 7. Try it yourself

```sh
# the lab in its container (agent on 7171)
make lab
docker exec cka-sim-lab-1 docker ps                     # the nodes, one level down
docker exec cka-sim-lab-1 docker exec cka-sim-control-plane kubectl get nodes

# or the agent straight on your machine's Docker, beside the container
cd apps/lab && go build -o /tmp/agent ./cmd/agent
CKA_SIM_AGENT_ADDR=127.0.0.1:7272 CKA_SIM_CLUSTER=cka-sim-local \
  CKA_SIM_AGENT_TOKEN=local-only-agent-token-not-a-secret /tmp/agent
T=local-only-agent-token-not-a-secret
curl -s -H "Authorization: Bearer $T" -H 'Content-Type: application/json' -d '{}' \
  http://127.0.0.1:7272/lab.v1.LabService/CreateCluster
docker ps                                               # cka-sim-local-control-plane and its workers

# a real cluster through the agent, with a shell running kubectl (minutes)
make lab-e2e
```
