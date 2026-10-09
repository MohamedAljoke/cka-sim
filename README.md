# cka-sim — a local CKA exam simulator

Practise for the **Certified Kubernetes Administrator** exam in an environment shaped like the
real one, on your own machine:

- You land on a host called **`base`** that has **no `kubectl`**.
- Every task tells you which host to use: **`ssh <host>`**, do the work, `exit` back to base.
- Task hosts are **real cluster nodes** with `kubectl`, the `k` alias and completion, `yq`,
  `curl`, `wget`, `man`, `vim`, `crictl`, `etcdctl`/`etcdutl`, and `sudo -i` for root —
  systemd kubelet, static Pod manifests and etcd included, so troubleshooting is real.
- **Kubernetes v1.35**, the exam's current version, across two clusters.
- A **2-hour timed exam panel** in your browser, laid out like the exam: questions with their
  ssh infobox, flags and allowed-docs links on the left, a **terminal on `base`** on the right
  (tabs, Ctrl+Shift+C / Ctrl+Shift+V, shells that survive a page reload). **End exam** grades every task against the live clusters with
  **partial credit**; pass mark **66%**.
- Afterwards, every question gets its check-by-check result, an **explanation of the
  concept** and a **reference solution**.

```
candidate@base:~$ kubectl get nodes
bash: kubectl: command not found
candidate@base:~$ ssh cka3962-worker
candidate@cka3962-worker:~$ k get nodes
NAME                    STATUS     ROLES           AGE   VERSION
cka3962-control-plane   Ready      control-plane   12m   v1.35.8
cka3962-worker          NotReady   <none>          12m   v1.35.8
candidate@cka3962-worker:~$ sudo -i
root@cka3962-worker:~# systemctl status kubelet
```

## Requirements

Linux or WSL2 with **Docker**, **kind** (≥ 0.24), **kubectl** and **Go** (≥ 1.25).
About **5 GB of free RAM** while the clusters run.

```sh
go install github.com/MohamedAljoke/cka-sim/cmd/cka-sim@latest
# or, from a clone:  go build -o bin/cka-sim ./cmd/cka-sim
```

## Use

From a clone, `make` lists shortcuts for the whole cycle:

```sh
make reset                 # drop everything and build it again
make exam MINUTES=60       # the panel (or: make study TASKS="tr-kubelet")
make open                  # in another terminal: open it — questions left, your shell on base right
make down                  # done for the day
```

Or with the binary directly:

```sh
cka-sim up                 # once: build images, create clusters cka7491 + cka3962 and base (~5 min)

cka-sim exam               # rebuild clean clusters, draw 16 tasks, set them up, start the clock
                           #   exam panel: http://localhost:8080
                           #   its terminal pane is your shell on base: candidate@base
cka-sim shell              # or the same shell in your own terminal

cka-sim study              # the same panel with no clock: check, reset and read each task's
                           #   solution as you go (study tr-kubelet ar-etcd picks tasks)

cka-sim list               # every task, by curriculum domain
cka-sim practice tr-kubelet   # one task, untimed
cka-sim check tr-kubelet      # grade it now
cka-sim solution tr-kubelet   # explanation + reference solution

cka-sim status             # what is running in docker, its memory, the current session
cka-sim down               # delete everything
```

`cka-sim exam -n 8 -minutes 60` makes a shorter mock exam. `-fresh=false` reuses the current
clusters instead of rebuilding them; `-resume` re-opens the panel of an exam in progress.

**Study mode** (`cka-sim study`) is for learning rather than testing yourself. The timer counts
up instead of down, and every question gets **Check my work** (grade just that task, as often
as you like), **Reset task** (run its setup again to retry from scratch), and the explanation
and reference solution, there whenever you want them. **Finish** grades everything and shows
the same report as an exam. Flags go before task ids: `cka-sim study -fresh=false tr-service`.

Want to understand how it works, or run it on a server? Follow
[docs/STUDY-GUIDE.md](docs/STUDY-GUIDE.md) step by step.

## Tasks

The draw follows the CKA v1.35 curriculum weights: Troubleshooting 30%, Cluster Architecture 25%,
Services & Networking 20%, Workloads & Scheduling 15%, Storage 10%.

| Domain | Task | You practise |
|---|---|---|
| Troubleshooting | `tr-kubelet` | node NotReady → kubelet systemd unit, `journalctl`, `daemon-reload`, enable on boot |
| Troubleshooting | `tr-scheduler` | Pods Pending → broken static Pod manifest of kube-scheduler |
| Troubleshooting | `tr-service` | Service with no endpoints and the wrong targetPort |
| Architecture | `ar-rbac` | ServiceAccount + Role + RoleBinding, least privilege, `auth can-i --as` |
| Architecture | `ar-etcd` | etcd snapshot save and restore — including the API-server step most guides skip |
| Networking | `nw-netpol` | NetworkPolicy with OR-ed pod and namespace selectors |
| Workloads | `wl-rollout` | rollback, revision numbers, rolling-update strategy |
| Storage | `st-pvc` | PersistentVolume, claim and Pod with matching storage classes |

More tasks are on the way: Helm, Kustomize, Gateway API, Ingress, CoreDNS, HPA, CRDs and
operators, kubeadm node join, scheduling with taints and affinity, ConfigMaps and Secrets,
logs and sidecar output, `kubectl top`.

## How it differs from the real exam

Honest list, so you know where the simulator is softer or harder:

- **No remote desktop and no proctoring.** Keep yourself to the docs the exam allows
  (kubernetes.io/docs, kubernetes.io/blog, github.com/kubernetes) — the panel links them.
- **Hosts are kind nodes (containers).** Everything on a node works — kubelet, static Pods,
  etcd, crictl — but you cannot reboot one, and a highly-available control plane or a
  from-scratch `kubeadm init` are not reproduced.
- **Calico is the CNI**, so NetworkPolicies are enforced. (kind's default CNI cannot enforce
  them on WSL2 kernels.)
- **Grading checks outcomes**, like the real exam: what exists and works, not which commands
  you typed.

## Add a task

A task is a directory `tasks/<domain>/<id>/`:

| File | Purpose |
|---|---|
| `task.md` | frontmatter (`id`, `title`, `domain`, `weight`, `cluster`, `host`) + **Context** and **Task**, written like an exam question |
| `setup.sh` | puts the cluster into its starting (often broken) state; must be re-runnable |
| `check.sh` | prints `PASS <points> <what>` / `FAIL <points> <what>` lines via the `check` helper |
| `solution.sh` | the reference fix, with the diagnosis as comments |
| `explain.md` | the concept behind the task — shown after the exam |

Scripts run on your machine with the helpers in [`tasks/lib.sh`](tasks/lib.sh)
(`k`, `on_host`, `check`, `wait_for`, `fresh_ns`, `remember`/`recall`).

```sh
cka-sim selftest <id>      # setup → must fail its checks → solution → must score full marks
make check                 # what CI runs: gofmt, vet, go test -race, bash -n on every script
make build | install | selftest | status
```

## Which docker containers are mine?

kind creates the cluster nodes itself, so there is no compose file: `cka-sim status` lists
everything the simulator owns, and since every container is named `cka…`,
`docker ps -a --filter name=cka` shows the same set. `cka-sim down` removes all of it.

## WSL2 notes

`cka-sim` handles three WSL2 quirks for you: Kubernetes 1.35 refuses cgroup v1 (the kubelet is
started with `failCgroupV1: false`), kind's default CNI cannot enforce NetworkPolicy without
`CONFIG_NFT_QUEUE` (Calico is used instead), and a leftover Docker Desktop credential helper
in `~/.docker/config.json` cannot block image pulls (docker runs with a private, empty config).
