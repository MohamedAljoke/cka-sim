# cka-sim

A practice simulator for the CKA exam: pick a task, get a real Kubernetes cluster, solve it in a terminal in
your browser, the way the exam does: start on a host, `ssh` to a node.

Not a mock. Every cluster is a real kubeadm cluster with a real etcd, static pods and kubelets, so you can
break a control plane and fix it the way the exam asks you to.

## Run it

You need Docker, with at least 2 CPUs and 4 GiB of memory for it. No clone needed:

```sh
curl -fsSL https://raw.githubusercontent.com/MohamedAljoke/cka-sim/rebuild/infra/install.sh | sh
```

That installs the `cka-sim` command in `~/.local/bin` and starts it. Open http://localhost:7080, pick a task
and click **Start lab**. The first cluster takes a few minutes while the node image downloads; later ones
reuse it.

```sh
cka-sim down        # stop it; the downloaded images stay
cka-sim up          # start it again
cka-sim status      # what is running
cka-sim logs lab    # follow one part's logs: lab or app
cka-sim shell       # the page's terminal, in yours
cka-sim update      # get the latest cka-sim and restart it
cka-sim uninstall   # delete it and everything it stored
```

From a clone, `make up` builds the same images from source and runs them.

To work on the code, `cp .env.example .env` and run `make` to list the commands: `make lab` starts only the
lab, and `make dev` runs the page with hot reload on http://localhost:5173.

## How it works

![The lab container runs its own dockerd, which runs the base host and the kind nodes](docs/img/lab-container.svg)

- **A whole cluster in one container.** The lab container runs its own Docker, and [kind](https://kind.sigs.k8s.io)
  turns each node into a container inside it. Three levels deep: your Docker → the lab's Docker → the cluster's
  containerd.
- **kind as a Go library, not a CLI.** The agent drives cluster creation in-process and turns kind's logger
  calls into live progress, with no log scraping. Config is embedded in the binary, and a cgroup v1 patch is
  added only when the host needs it.
- **An exam-like base host.** The terminal lands on a host with no kubectl, as `candidate`. Each task says
  which node to `ssh` to, and every node has candidate's key, a kubeconfig and passwordless `sudo`. The node
  image adds sshd, vim and `k` completion on top of `kindest/node`.
- **A browser terminal that is a real shell.** Keystrokes travel WebSocket → platform → a bidirectional Connect
  stream → a pty on the agent → `docker exec -it` → bash on the base host. What you see is bash's echo coming
  back; resizes are copied down, through `ssh` too, so `vim` and `top` draw correctly.
- **The platform never touches Docker.** It only knows the agent's contract
  ([`lab.proto`](apps/go-shared/contract/lab/v1/lab.proto)), so a lab could later run on a VM or GKE without the
  platform changing.
- **Clean ends.** Killing `docker exec` leaves bash running on the base host, so each shell is tagged in its
  environment and the agent hunts down every tagged process on close.

![One keystroke from xterm.js to bash on the base host and back](docs/img/terminal.svg)

The full walkthrough, from Docker-in-Docker quirks to the base host's ssh setup and how HTTP/2 streams both ways, is in
[docs/LAB-INTERNALS.md](docs/LAB-INTERNALS.md).

## The code

| Path | What it is |
| --- | --- |
| `apps/platform` | Go API: task catalog (`task.md`, `setup.sh`, `check.sh` per task) and the terminal proxy |
| `apps/lab` | Go agent that runs inside the lab container: cluster lifecycle and shells |
| `apps/web` | TypeScript + Vite page with xterm.js |
| `apps/go-shared` | Protobuf contracts and shared helpers, the only code apps may share |
| `infra` | The images, the one Compose file that runs them, and the `cka-sim` command with its installer |

Apps never import each other, each follows a domain / use cases / adapters layout, and every endpoint has
end-to-end tests against the real binary. The rules are in [AGENTS.md](AGENTS.md).

**Stack:** Go, Connect + Protobuf, kind, Docker-in-Docker, creack/pty, TypeScript, Vite, xterm.js.

## Status

Being rebuilt step by step on this branch; [docs/REBUILD-STATUS.md](docs/REBUILD-STATUS.md) tracks what has
landed and [docs/REBUILD-IN-STEPS.md](docs/REBUILD-IN-STEPS.md) what comes next. Next up: a killer.sh-style exam
mode.
