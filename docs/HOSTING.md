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
| Terminal in the browser                | `apps/web` xterm → websocket → `internal/server` → shell in node |
| Nodes with systemd + kubelet           | kind node image (`internal/cluster/node.Dockerfile`)          |

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

## Next step: test before designing

1. **Fly Machine.** Create one 4 GB Machine from an image with dockerd + kind, run `kind create cluster` with our
   `kind.yaml`, then run `wl-scale`'s `setup.sh` and `check.sh`. Time each step.
2. **Sprite.** Try the same. If kind works *and* checkpoint/restore brings the cluster back healthy, Sprites
   become a serious option.

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
