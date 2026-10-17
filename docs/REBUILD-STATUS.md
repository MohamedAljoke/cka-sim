# Rebuild status

Where the `rebuild` branch stands against [REBUILD-IN-STEPS.md](REBUILD-IN-STEPS.md), the plan brought over
from `main`. The plan describes `main`'s code; on this branch [AGENTS.md](../AGENTS.md) wins wherever the two
disagree. Its links to `BUILD-FROM-ZERO.md` and `IMPROVEMENTS-*.md` point at files that only exist on `main`.

Out of scope here: Part 3 (the assistant, the eval harness, docs RAG) and 4.3 (GKE).

Update this file when a step lands, and the [cka-sim docs](https://claude.ai/code/artifact/5f47a2bd-5b68-4d15-b90f-1d76ae14750d)
when what they describe changes.

## Where each step stands

| Step | Status | On `rebuild` |
| --- | --- | --- |
| 0.1 Skeleton and go-shared | Done | Plus `rpcauth` and `contract/` (`.proto` + Connect, `make proto`), instead of copying contracts into each app |
| 1.1 Task format and loader | Done | All 202 tasks, `task.md` + `explain.md` only |
| 1.2 HTTP API | Done | `GET /api/tasks`, `/api/tasks/{id}`, `/api/tasks/{id}/solution` |
| 1.3 The page | Done | |
| 1.4 Filters and the catalog | Partly | Filters done; `docs/CKA-QUESTION-BANK.md` and the `new-task` skill not brought |
| 1.5 Accounts | Not started | `apps/auth`, Postgres, session JWT, the platform's verifier, login page |
| 2.1 `lab doctor` | Not started | No `lab` CLI at all yet (`doctor`, `up`, `down`, `shell`, `version`) |
| 2.2 The lab image | Partly | Same image shape, but the agent builds the cluster on `CreateCluster` instead of `lab up` at start |
| 2.3 Exam-like nodes and base host | Done | The kind adapter builds `cka-sim/node` (`candidate` with sudo, sshd, `k`) and `cka-sim/base` (no kubectl) inside the lab's Docker; `<cluster>-base` joins the `kind` network, its key and the admin kubeconfig go to every node, and root gets a kubeconfig on every node for task scripts. `OpenShell` lands on base as `candidate`; `cka-sim shell` too. No image preload from `catalog/images.txt` yet |
| 2.4 The agent | Done | `/healthz`, token, `LabService` Get/Create/Delete, `OpenShell` (a bidi stream; `docker exec` under a pty, tagged shells hung up on close) and `RunScript` (bash as root on a node; exit code and the last 64 KiB of output; a deadline kills the script's whole process group) |
| 2.5 The platform reaches the lab | Partly | `labagent` adapter, `/api/lab`, polling page; no sandbox `Provider`, desks or idle end |
| 2.6 Terminal in the page | Done | `/ws/terminal` bridges the page's websocket to `OpenShell`; the shell is `candidate` on the base host, so the page's `ssh <host>` hint is the first command |
| 2.7 Setup, check, solution | Partly | `lib.sh`, every `setup.sh` and `check.sh` are in the catalog. `/api/tasks/{id}/start` runs setup and `/api/tasks/{id}/check` grades the check's PASS/FAIL lines (`tasks.ReadGrade`), both through `RunScript`; the page has Set up task and Check. Tasks that need an add-on (2.11) fail for now. No `solution.sh` yet (the selftest, 2.8, needs it) |
| 2.8 Selftest | Not started | |
| 2.9 Tasks that break the cluster | Not started | `reset.sh` (25 tasks on `main`), `Heal`, `Tidy` |
| 2.10 The exam | Not started | `Draw`, exam session saved to a file, weighted score, `/api/exam*`, exam page |
| 2.11 Add-ons | Not started | metrics-server, kube-router, Gateway API + Traefik, VPA, registry, chart repo |
| 4.1 One command, Docker only | Partly | `install.sh` puts the `cka-sim` command (`up`, `down`, `status`, `logs`, `update`, `uninstall`) and `infra/compose.yml` in `~/.cka-sim`; it runs the lab and the app (the page and the platform behind Caddy) on :7080 with a token generated once. `cka-sim shell` opens the base host. No auth or assistant yet |
| 4.2 Published images | Partly | `images.yml` pushes `cka-sim-lab` and `cka-sim-app` (amd64 + arm64) to GHCR as `:rebuild` and `:sha-…` on every push to `rebuild`; no `ci.yml` yet |

## Next

A task is playable exam-style: the terminal sits on the base host, `ssh <host>` reaches the task's node, set it
up, solve it, Check. Next is exam mode (2.10): draw N tasks by domain weight, set them all up front, a timer,
and a score at the end. Selftest (2.8) and reset (2.9) make that safe for tasks that break the cluster, and
add-ons (2.11) come in with the first task that needs each one.
