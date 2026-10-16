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
| 2.4 The agent | Partly | `/healthz`, token, `LabService` Get/Create/Delete; no script runner, no shell |
| 2.5 The platform reaches the lab | Partly | `labagent` adapter, `/api/lab`, polling page; no sandbox `Provider`, desks or idle end |
| 2.6 Terminal in the page | Not started | `/ws/terminal` proxied to the agent, xterm.js |
| 2.7 Setup, check, solution | Not started | `setup.sh`, `check.sh`, `solution.sh`, `lib.sh`, `grader`, `/start` and `/check`, Check button |
| 2.8 Selftest | Not started | |
| 2.9 Tasks that break the cluster | Not started | `reset.sh` (25 tasks on `main`), `Heal`, `Tidy` |
| 2.10 The exam | Not started | `Draw`, exam session saved to a file, weighted score, `/api/exam*`, exam page |
| 2.11 Add-ons | Not started | metrics-server, kube-router, Gateway API + Traefik, VPA, registry, chart repo |
| 4.1 One command, Docker only | Partly | `infra/local/compose.yml` runs the lab only; no `cka-sim` command or `install.sh` |
| 4.2 Published images | Not started | `images.yml`; `main` also has `ci.yml` testing lab and platform on Linux, macOS and Windows |

## Next

Steps 2.3, 2.4's runner and shell, 2.6 and 2.7 together make one task playable end to end: open it, get a
terminal on its host, run setup, solve, Check. On this branch that means the `LabService` contract grows a
script-running RPC and a terminal stream beside it. Exam mode (2.8 to 2.10) builds on that, and add-ons (2.11)
come in with the first task that needs each one.
