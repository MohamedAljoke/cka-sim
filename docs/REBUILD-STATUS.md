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
| 2.3 Exam-like nodes and base host | Not started | `node.Dockerfile`, `base.Dockerfile`, ssh between them, image preload from `catalog/images.txt` |
| 2.4 The agent | Done | `/healthz`, token, `LabService` Get/Create/Delete, `OpenShell` (a bidi stream; `docker exec` under a pty, tagged shells hung up on close) and `RunScript` (bash as root on a node; exit code and the last 64 KiB of output; a deadline kills the script's whole process group) |
| 2.5 The platform reaches the lab | Partly | `labagent` adapter, `/api/lab`, polling page; no sandbox `Provider`, desks or idle end |
| 2.6 Terminal in the page | Done | `/ws/terminal` bridges the page's websocket to `OpenShell`; the shell is root on the control plane until 2.3 brings the base host, so the page's `ssh <host>` hint doesn't apply yet |
| 2.7 Setup, check, solution | Partly | `lib.sh` and every `setup.sh` are in the catalog, `/api/tasks/{id}/start` runs setup through `RunScript`, and the page has Set up task; tasks that need an add-on (2.11) or a kubeconfig on a worker (2.3) fail their setup for now. No `check.sh`, `solution.sh`, `grader`, `/check` or Check button yet |
| 2.8 Selftest | Not started | |
| 2.9 Tasks that break the cluster | Not started | `reset.sh` (25 tasks on `main`), `Heal`, `Tidy` |
| 2.10 The exam | Not started | `Draw`, exam session saved to a file, weighted score, `/api/exam*`, exam page |
| 2.11 Add-ons | Not started | metrics-server, kube-router, Gateway API + Traefik, VPA, registry, chart repo |
| 4.1 One command, Docker only | Partly | `infra/local/compose.yml` runs the lab only; no `cka-sim` command or `install.sh` |
| 4.2 Published images | Not started | `images.yml`; `main` also has `ci.yml` testing lab and platform on Linux, macOS and Windows |

## Next

The rest of 2.7 makes a task playable end to end: set it up, solve it, Check. `check.sh` runs through
`RunScript` like setup does, so it needs no lab change. Then 2.3 moves the shell from the control plane to the
base host. Exam mode (2.8 to 2.10) builds on that, and add-ons (2.11)
come in with the first task that needs each one.
