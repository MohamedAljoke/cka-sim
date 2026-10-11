# Hosting cka-sim for other users

Today cka-sim runs one local kind cluster for one user. This document covers what it takes to turn it
into a hosted, KodeKloud-style lab site. The focus is the hard part: **a separate cluster for each user**.

## The lab model stays the same

KodeKloud, Killercoda and killer.sh appear to follow the same pattern we already use:
**set up a known state → user works on a real system → run checks against the live cluster → show the result**.

| Lab platform (as seen from outside)   | cka-sim                                                      |
|----------------------------------------|--------------------------------------------------------------|
| Lab/question definition                | `catalog/<task>/task.md` (front matter + text)                |
| Environment prepared before a question | `setup.sh`                                                    |
| "Check" button / end-of-exam grading   | `check.sh` → `PASS <pts> ...` / `FAIL <pts> ...`             |
| Solution / explanation tab             | `solution.sh` + `explain.md`                                  |
| Terminal in the browser                | `apps/web` xterm → websocket → `apps/platform/internal/server` → lab agent → shell on base |
| Nodes with systemd + kubelet           | kind node image (`apps/lab/internal/cluster/node.Dockerfile`) |

What we'd need to add is the multi-user infrastructure around this model, not a different lab model.

## What changes when it's hosted

- **Accounts, sessions, scores, payments.** These are ordinary web-app work.
- **Terminal gateway.** The websocket checks the user's session and routes the connection to *that user's*
  sandbox, instead of a fixed local `docker exec`.
- **One isolated cluster per user.** This is created on demand, has a hard time limit, and is destroyed afterwards.
  It is the core problem and the subject of the rest of this document.

## Per-user cluster options

| Option | How it works | Pros | Cons |
|---|---|---|---|
| 1. VM per user + kind | Each user gets a VM running our current kind setup | Almost no code changes; the VM is the security boundary | ~1–2 min boot, a whole VM (~4 GB) per user |
| 2. VM per node + kubeadm | `controlplane` and `node01` are separate VMs from a prepared image | Most exam-realistic (SSH between nodes, `kubeadm upgrade`, etcd backup) | 2× machines, slower, networking between nodes to manage |
| 3. Micro-VMs on big hosts | Firecracker/Kata on bare metal, many sandboxes per host | Fast boot and cheap per user | A lot of infrastructure work |

To avoid:
- **Shared host with kind / Docker-in-Docker.** Privileged containers let one user escape to the host.
- **vcluster / namespaces on a shared cluster.** There are no nodes, kubelet or etcd to break, which removes about half the CKA syllabus.

## Design points (any option)

- **Golden image.** Bake in everything ahead of time: the node image, kind/kubeadm, the catalog and pre-pulled
  images (nginx, ...). Startup becomes just "boot + create cluster".
- **Warm pool.** Keep N ready clusters so "Start lab" is instant, and refill the pool in the background.
- **Reset vs. recreate.** `fresh_ns` is enough between tasks within one session. Between users, always destroy
  the cluster and never reuse it.
- **Lifecycle state machine.** Track each sandbox through `provisioning → ready → assigned → expired → destroying`.
  A reaper job kills expired sandboxes and anything the database doesn't know about (leaked VMs cost money).
- **Lockdown.** Restrict outbound network (block mining pools, spam, scanning), set hard CPU/RAM/time limits, and give
  sandboxes no cloud credentials.
- **Reaching the sandbox.** The engine needs a shell (terminal) and a way to run a script (setup/check).
  SSH or a tiny agent in the VM provides both. These plug into the existing swappable `Opener` and
  script-runner interfaces.

## Fly.io (we have credits)

### Fly Machines: recommended

- Firecracker VMs with their own kernel; our process is PID 1 with full root, so `dockerd` + kind can run inside
  it like they do locally. The VM isolates each user.
- Matches the golden-image idea: one OCI image containing dockerd, kind, the pre-loaded node image and the catalog.
- REST API: `POST /v1/apps/{app}/machines` with `guest: {cpus, memory_mb}`. Start/stop, `auto_destroy` and
  per-Machine metadata cover most of the pool and reaper.
- The private network (6PN) lets the backend reach each Machine without exposing it publicly.
- Catches:
  - The root filesystem resets on every restart (fine for throwaway labs).
  - Each Machine needs about 4 GB of RAM for kind.
  - `kind create` takes about 1 min unless we keep warm Machines ready.
  - Fly's docs don't officially cover running dockerd in a Machine, so we must test it ourselves.

### Fly Sprites: interesting, but risky for us

- Appealing parts:
  - **Checkpoint/restore** ("build the cluster once, restore it per user or task").
  - Per-second billing with compute free while idle.
  - Hardware isolation, 8 vCPU / 100 GB.
- Concerns:
  - **No custom image.** Every Sprite starts from a fixed Ubuntu base that we install things into.
  - **Restricted capabilities.** Only 16 of 41 are available, and `--privileged` doesn't change that. Users report that `docker exec` fails for
    containers that drop root. kind nodes are privileged systemd containers, so this is a likely blocker. A missing
    `SYS_PTRACE` also hurts debugging tasks.
  - **Hibernation.** After about 30s idle a Sprite suspends, and when "cold" its processes are dropped. A kubelet/etcd
    cluster that stops while a user reads a task is bad. The Tasks API can keep a Sprite awake, but that goes against how Sprites are meant to be used.
  - Sprites are built for AI agents and dev environments, not nested container platforms.

## GKE as the fleet manager (we have GCP credit)

The idea is to learn GKE by letting it run the fleet of labs, in place of the Fly pool. It is not a GKE
cluster per user: a managed cluster hides etcd, the kubelet and static pods, the same problem as vcluster.

### Shape

```
GKE Standard cluster
├─ node pool "platform" → main app (Go server, websocket proxy)
└─ node pool "labs"     → one lab pod per node (privileged: dockerd + kind + agent)
                          tainted so only labs land there; the cluster autoscaler adds and removes nodes
```

- **One lab per node.** GKE nodes are Compute Engine VMs, so the VM stays the boundary between users,
  as in option 1. GKE only decides where that VM comes from.
- **The pool is [agent-sandbox](https://github.com/kubernetes-sigs/agent-sandbox)** (Kubernetes SIG, supported on GKE):
  - `SandboxTemplate` is the lab pod spec.
  - `SandboxWarmPool` with `replicas: N` keeps N labs up.
  - `SandboxClaim` checks one out; deleting the claim ends the lab.

  This replaces what the Fly pool does by hand.
- **Code change.** A `gke` provider behind `sandbox.Provider` creates a claim, waits for Ready, and talks to the
  pod's agent. The agent and the per-user workspace are needed on any platform.

### Catches

- **The lab pod must be privileged** to run kind, and a privileged pod owns its node. That is only acceptable
  because the node holds one user:
  - Enforce one lab per node with a taint, anti-affinity, and requests that fill the node.
  - **Delete the node after each user**, not just the pod.
- **Standard, not Autopilot.** Autopilot restricts privileged pods. agent-sandbox's examples isolate with gVisor
  or Kata. gVisor likely can't run kind (nested Docker with systemd), and Kata needs nested virtualization. So
  the plan is plain runc on a dedicated node.
- **The network wall lives outside the node.** Root on a node can switch off NetworkPolicy there, so the rules
  that matter are VPC firewall rules on the lab nodes' tag: no lab node reaches another, the control plane or
  the metadata server. In addition:
  - Turn on Workload Identity.
  - Give the lab nodes a service account with no roles.
- **No suspend.** A warm lab is a running node billed by the hour, where Fly parks a suspended VM for its disk
  only. Keep the pool at 1–2, or at 0 and accept a cold start (new node ~1–2 min + `kind create` ~1 min).
- **Credit runs out.** The GCP free trial credit lasts 90 days (ending around early December 2026). The free
  tier covers one zonal cluster's management fee; nodes are billed.

### Tested on GKE (2026-10)

The kit was `apps/cka-sim/deploy/gke`; since the platform/lab split it is `infra/gke` (`make` lists the steps), with
the lab image in `apps/lab/deploy` and the app in `apps/platform/deploy/gke`. It used its own project, `cka-sim-gke-maljoke`,
so `make down` removes everything. The cluster was zonal in `southamerica-east1-a`, with a system pool of one
`e2-medium` and a labs pool of `e2-standard-2` Ubuntu nodes, autoscaled from 0 to 2. agent-sandbox was v1.0.4.

| Step | Time |
|---|---|
| New lab node from 0 (autoscaler scale-up until the pod is scheduled) | ~3.5 min |
| Lab pod's `cka-sim up` on a fresh node (builds the node image, creates kind) | 2–4 min |
| **Claim a warm lab** (`SandboxClaim` → Ready) | **1s** |
| `wl-scale` setup / check / solution / check | 3s / 1s / 3s / 1s |

What we learned:
- **kind runs with no workarounds**, unlike Fly's three. An `emptyDir` at `/var/lib/docker` gives `overlay2`,
  and the Ubuntu node's cgroup v2 works as-is. `tr-kubelet` (systemd inside the kind nodes) passes too.
- **Requests must fit next to GKE's own daemons.** On an `e2-standard-2`, a request of 1.5 CPU and 5 GiB never
  triggered a scale-up. 1 CPU and 4 GiB does.
- **The pod builds its own cluster at boot** (`lab-start.sh`). A readiness probe asks the kind API server for
  `/readyz`, so a warm lab is one whose cluster answers.
- **agent-sandbox adds a NetworkPolicy by default.** It allows outbound traffic to the internet, but not to private
  ranges or `169.254.0.0/16`, and inbound traffic only from its router. That blocks lab pod → lab pod. It did
  **not** block a lab reaching another lab *node's* kubelet (`:10250`), because Dataplane V2 treats node
  addresses as nodes, not as the private range.
- **The VPC firewall closes that gap, but only with a separate pod range for labs.** A deny on the whole
  cluster pod range also blocked GKE's konnectivity agent, which the control plane uses to reach each kubelet,
  and `kubectl exec` broke. So the labs pool gets its own range (`10.80.0.0/16`), and the rule
  (`make wall`) denies the lab nodes' tag and that range.
- Result: lab A, from its pod and from inside its kind node, can't reach lab B's pod or node. `kubectl exec` and
  tasks still work.

### The app and the lab on separate machines (2026-10)

The server now runs on the system pool with `CKA_SIM_PROVIDER=gke`. It claims a warm lab and reaches it
through the agent in the lab pod (flow 7 in `docs/FLOWS.md`). On the site, a task start shows "Lab ready in
1.4s · GKE". Setup, check, the terminal (`candidate@base`, `ssh` to the nodes) and **End lab** all work. The
app pod has no Docker. The app can reach a lab's agent; another lab can't, from either its pod or its kind node.

What we learned:
- **The readiness probe must be the agent's `/healthz`.** Probing the kind API server passed while
  `cka-sim up` was still joining workers, so the pool handed out labs whose agent wasn't listening yet.
- **DNS needs its own egress rule.** GKE turns on NodeLocal DNSCache, so a pod's DNS goes to `node-local-dns` on
  its node, in a private range. agent-sandbox's default egress blocks it, and image pulls in `cka-sim up` time
  out. The template allows `node-local-dns` and `kube-dns` on port 53.
- **A template change doesn't reach warm labs by default.** Each Sandbox keeps its own copy of the pod spec.
  The pool's `updateStrategy: Recreate` rebuilds the warm ones.
- **The system pool needs room.** One `e2-medium` is fully booked by GKE's pods and the agent-sandbox
  controller. It now autoscales from 1 to 2 nodes.

Not checked yet:
- **What a lab can get from the metadata server.** It answers from inside a kind node with the Workload Identity
  pool name. Whether it hands out a token, and what that token can do, is untested.
- **The GKE API's public endpoint is reachable** from a lab. It needs credentials, but a private endpoint or
  authorized networks would hide it.
- **The nodes have public IPs**, and the default network allows SSH from anywhere. Private nodes would remove both.
- **Outbound internet is open.** That's fine for image pulls, but a registry allow-list would stop abuse.
- **Node reuse.** When a claim ends, the pool's next lab can land on the node the last user had. Each released node
  must be cordoned and deleted (or the pod made to fill a node that is then drained).

## Next step: test before designing

1. **Fly Machine.** Create one 4 GB Machine from an image with dockerd + kind, run `kind create cluster` with our
   `kind.yaml`, then run `wl-scale`'s `setup.sh` and `check.sh`. Time each step. *Done on the `fly` branch:
   kind runs, and a suspended VM resumes with a healthy cluster in ~1.4s.*
2. **Sprite.** Try the same. If kind works *and* checkpoint/restore brings the cluster back healthy, Sprites
   become a serious option.
3. **GKE.** In order, each one gating the next:
   1. Create a Standard cluster with one `e2-standard-2` node on the Ubuntu image. Run a privileged pod with
      dockerd + kind + our node image, then `wl-scale`'s setup and check. Time each step.
   2. Install agent-sandbox, create a template and a warm pool of 1, claim a lab, and time it.
   3. From inside lab A, try to reach lab B and the metadata server. Both must fail.

   *Done: all three pass, except that metadata server access is not yet checked. See "Tested on GKE".*

If Machines work, the code change is small: a `FlyProvider` behind the cluster-provider interface, plus an
opener and script runner that reach the Machine over the private network instead of local `docker exec`.

## Sources

- [Fly Docs: Machines API, create a Machine](https://docs.fly.io/machines/api/machines-resource)
- [Fly Docs: Working with Docker on Fly.io](https://docs.fly.io/blueprints/working-with-docker)
- [Fly Docs: Sprites overview](https://docs.fly.io/sprites)
- [Fly Docs: Sprite lifecycle](https://docs.fly.io/sprites/concepts/lifecycle)
- [Fly Docs: Sprite checkpoints](https://docs.fly.io/sprites/concepts/checkpoints)
- [Fly community: docker exec doesn't work for most images within a sprite](https://community.fly.io/t/docker-exec-doesnt-work-for-most-images-within-a-sprite/27956)
- [Fly community: How to get Docker running on Sprites?](https://community.fly.io/t/how-to-get-docker-running-on-sprites/27168)
- [The Design & Implementation of Sprites (Fly blog)](https://fly.io/blog/design-and-implementation/)
- [agent-sandbox: Sandbox, SandboxTemplate, SandboxWarmPool, SandboxClaim](https://github.com/kubernetes-sigs/agent-sandbox)
