# Status

What has been built, step by step, under the original plan's step numbers. What's left is in
[BUILD-IN-STEPS.md](BUILD-IN-STEPS.md) and [ASSISTANT.md](ASSISTANT.md).

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
| 2.5 The platform reaches the lab | Done | `labagent.Lab` talks to the agent of a `Machine`: `labagent.Fixed` (one URL, locally) or `kube.Pod` (a pod made on Start lab); `/api/lab`, polling page. `labagent.Fleet` gives each student their own `Lab` (3.2) |
| 2.6 Terminal in the page | Done | `/ws/terminal` bridges the page's websocket to `OpenShell`; the shell is `candidate` on the base host, so the page's `ssh <host>` hint is the first command |
| 2.7 Setup, check, solution | Partly | `lib.sh`, every `setup.sh` and `check.sh` are in the catalog. `/api/tasks/{id}/start` runs setup and `/api/tasks/{id}/check` grades the check's PASS/FAIL lines (`tasks.ReadGrade`), both through `RunScript`; the page has Set up task and Check. Every task has a `solution.sh` |
| 2.8 Selftest | Done | `tasks/usecases.SelftestTask`: start (heal + setup), check earns 0, `solution.sh`, check earns full. `make selftest [TASK=<id>]` runs `catalog/selftest_test.go` behind `-selftest`. 198 of 203 pass: `tr-events` earns 3/4 after its solution, and the four `wl-hpa-*` checks outlast the 2-minute check timeout |
| 2.9 Tasks that break the cluster | Done | 25 tasks have a `reset.sh` that does nothing on a healthy cluster. A setup marks its task touched; `HealCluster` runs the touched tasks' resets one at a time before every practice start and before an exam's setups, and a failed reset is tried again next time. `TidyTask` (`POST /api/tasks/{id}/tidy`, sent by the page when it leaves a task it set up) resets the task and deletes the namespaces `fresh_ns` labelled. A scored exam tidies all its tasks, and the next exam waits for that. Setups share the cluster; a reset or tidy has it alone. The touched list lives in memory, so a restarted platform forgets what practice broke |
| 2.10 The exam | Done | `tasks.Draw` (domain weights, at most one `order: last` task, set up last and shown first); the `exams` domain with Begin, Get, Flag, End, Discard and Resume. Each student's exam is saved to `<CKA_SIM_STATE_DIR>/exams/<student>.json` (an older `exam.json` moves to `local.json`); setups and checks run in the background and a restart resumes them. The clock starts once every setup is done, and the server ends an exam read past its deadline. Score weighted by task, pass at 66%; a task whose setup failed counts in neither earned nor total. `/api/exam*`; practice start, check and solution answer 409 during an exam. The page at `#/exam`: form, setup progress, pills, flags, countdown, terminal, results. Healed before and tidied after (2.9) |
| 2.11 Add-ons | Done | The kind adapter installs them after the base host, from images in the lab's Docker: metrics-server, kube-router (firewall only) plus an nftables table that refuses Services without endpoints, Gateway API + Traefik (also serving the `nginx` IngressClass), the VPA recommender, helm on every node, a private registry (`cka-sim-registry:5000`, with auth) and a chart repo (`http://cka-sim-charts`); the web-operator is only staged for `ar-operator`. The controller manager marks a node NotReady after 20s. Any add-on that fails fails the whole build |
| 3.1 Students | Partly | `CKA_SIM_STUDENTS` is a fixed `name=token` list; `/api/join?token=` puts the token in the `cka_sim_student` HttpOnly cookie and redirects; other requests without a known cookie get 401 `student_unknown`, and the page asks for the link. Empty list: everyone is `local`. Not built: signed JWTs, `server mint`, expiry, the denylist |
| 3.2 A lab per student | Partly | `labagent.Fleet`; the use cases, touched set, script lock and exam are per owner; `kube.Pod` is `lab-<student>` labelled `cka-sim/owner`; the labs pool caps at `var.max_labs` (default 2); `make links`, `make labs`, `make lab-down NAME=`. Not built: the idle stop (and the Role's `list`), a "site is full" message |
| 4.1 One command, Docker only | Partly | `install.sh` puts the `cka-sim` command (`up`, `down`, `status`, `logs`, `update`, `uninstall`) and `infra/local/compose.yml` in `~/.cka-sim`; it runs the lab and the app (the page and the platform behind Caddy) on :7080 with a token generated once. `cka-sim shell` opens the base host. No auth or assistant yet |
| 4.2 Published images | Partly | `images.yml` pushes `cka-sim-lab` and `cka-sim-app` (amd64 + arm64) to GHCR as `:main` and `:sha-…` on every push to `main`; no `ci.yml` yet |
| 4.3 The hosted site on GKE | Partly | `infra/gke`: Terraform (zonal cluster, a system pool and a tainted labs pool scaling 0..`max_labs`, the site's IP) and kustomize manifests (the app with `CKA_SIM_LABS=kubernetes` and the students' Secret, RBAC on pods in `cka-sim-labs`, a NetworkPolicy, a Caddy IP allowlist on `<ip>.sslip.io`). `kube.Pod` and `labagent.Lab` have unit tests; `labagent.Fleet` has none |
| A1a TutorService | Done | `tutor/v1/tutor.proto`, served by the platform at `/tutor.v1.TutorService/` when `CKA_SIM_ASSISTANT_TOKEN` is set: `GetCurrentTask` (the task set up last and not tidied, per student, in memory), `CheckTask`, `InspectCluster` (`labs.ReadCommand`: an allowlist of kubectl read verbs and flags, quoted, `timeout 15`, run as root on the control plane, output cut to 16 KiB) and `GetSolution`. Unknown students are NOT_FOUND; every call is refused while that student's exam runs |
| A1b MCP server | Done | `apps/assistant`, its own module: `make assistant` serves MCP (go-sdk, Streamable HTTP) on `127.0.0.1:7373/mcp` for the student in `CKA_SIM_STUDENT` (`local`). Tools `current_task`, `run_check`, `inspect_cluster`, `explain` call `TutorService` with the token; `reviewer.md` is the server's instructions. A platform refusal reaches the AI with its reason; other failures are logged and answered without a cause |

## Next

Evals (Part 2 in [ASSISTANT.md](ASSISTANT.md)), then the rest of [BUILD-IN-STEPS.md](BUILD-IN-STEPS.md).
