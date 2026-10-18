# Status

Where cka-sim stands against [BUILD-IN-STEPS.md](BUILD-IN-STEPS.md). Its **Before** lines point at the first
version on the `archive/v2` branch; [AGENTS.md](../AGENTS.md) wins wherever the two disagree.

Out of scope for now: 4.3 (GKE).

Update this file when a step lands, and the [cka-sim docs](https://claude.ai/code/artifact/5f47a2bd-5b68-4d15-b90f-1d76ae14750d)
when what they describe changes.

## Where each step stands

| Step | Status | Notes |
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
| 2.7 Setup, check, solution | Partly | `lib.sh`, every `setup.sh` and `check.sh` are in the catalog. `/api/tasks/{id}/start` runs setup and `/api/tasks/{id}/check` grades the check's PASS/FAIL lines (`tasks.ReadGrade`), both through `RunScript`; the page has Set up task and Check. No `solution.sh` yet (the selftest, 2.8, needs it) |
| 2.8 Selftest | Not started | |
| 2.9 Tasks that break the cluster | Not started | `reset.sh` (25 tasks on `archive/v2`), `Heal`, `Tidy` |
| 2.10 The exam | Done | `tasks.Draw` (domain weights, at most one `order: last` task, set up last and shown first); the `exams` domain with Begin, Get, Flag, End, Discard and Resume. The exam is saved to `<CKA_SIM_STATE_DIR>/exam.json`; setups and checks run in the background and a restart resumes them. The clock starts once every setup is done, and the server ends an exam read past its deadline. Score weighted by task, pass at 66%; a task whose setup failed counts in neither earned nor total. `/api/exam*`; practice start, check and solution answer 409 during an exam. The page at `#/exam`: form, setup progress, pills, flags, countdown, terminal, results. No heal or tidy (2.9): restart the lab after an exam |
| 2.11 Add-ons | Done | The kind adapter installs them after the base host, from images in the lab's Docker: metrics-server, kube-router (firewall only) plus an nftables table that refuses Services without endpoints, Gateway API + Traefik (also serving the `nginx` IngressClass), the VPA recommender, helm on every node, a private registry (`cka-sim-registry:5000`, with auth) and a chart repo (`http://cka-sim-charts`); the web-operator is only staged for `ar-operator`. The controller manager marks a node NotReady after 20s. Any add-on that fails fails the whole build |
| 4.1 One command, Docker only | Partly | `install.sh` puts the `cka-sim` command (`up`, `down`, `status`, `logs`, `update`, `uninstall`) and `infra/compose.yml` in `~/.cka-sim`; it runs the lab and the app (the page and the platform behind Caddy) on :7080 with a token generated once. `cka-sim shell` opens the base host. No auth or assistant yet |
| 4.2 Published images | Partly | `images.yml` pushes `cka-sim-lab` and `cka-sim-app` (amd64 + arm64) to GHCR as `:main` and `:sha-…` on every push to `main`; no `ci.yml` yet |

## Next

Exam mode works on one shared cluster: draw, set everything up, play against the clock, End exam, a weighted
score. With the add-ons in (2.11), the gap left is that nothing resets the cluster between exams, so the lab needs
a restart after each one (2.9). Selftest (2.8) is the way to find out which tasks set up and grade cleanly.
