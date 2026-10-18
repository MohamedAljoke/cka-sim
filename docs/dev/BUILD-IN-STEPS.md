# Build cka-sim in steps: what's left

cka-sim is built one small feature at a time, and you **review each piece as it lands** instead of meeting a
finished system. One step is one feature: ask the AI to build it, check it yourself, commit, then move on. Each
step has:

- **Scope:** what the feature covers. It's what to ask for, not a file list.
- **Review:** what to try or read before you accept it.
- **Learn:** the idea to understand first, so you can review the code, not just run it.
- **Before:** where it lived in the first version, on `archive/v2`. Take its behaviour, never its structure:
  [AGENTS.md](../../AGENTS.md) wins.

The steps already built are in [STATUS.md](STATUS.md) and the git history. This file keeps only what's left, in
the suggested order; reorder freely.

## 1 · The assistant

An AI CLI in a second terminal in the page, in its own app (`apps/assistant`), with the platform as a proxy and
context on your task. Three parts, in [ASSISTANT.md](ASSISTANT.md).

## 2 · Finish the lab

### 2.1 The five tasks the selftest fails

- **Scope:**
  - `tr-events` earns 3 of 4 after its solution ("Scheduled comes before FailedMount").
  - The four `wl-hpa-*` checks take longer than the 2-minute check limit. Make the checks faster, or give a
    task its own limit.
- **Review:** `make selftest TASK=<id>` passes for each.

### 2.2 The hosted lab, live

- **Scope:** `make -C infra/gke up`, then fix what only a real cluster shows: kind in a privileged pod on
  Ubuntu nodes, the NetworkPolicy, the autoscaler's timing.
- **Review:**
  - Start lab shows "Waiting for a node", then the build steps.
  - A task sets up and checks.
  - Another IP gets no page.
  - Stop lab deletes the pod, and the node goes away.
- **Learn:** what offline validation can't prove.

### 2.3 Image preload

- **Scope:** pull the images the catalog's tasks use (`catalog/images.txt`) into the lab's Docker when it
  starts, and load them into the kind nodes, so a setup never waits on a registry.
- **Review:** a task's first setup doesn't pull.

### 2.4 `lab doctor`

- **Scope:** checks, in order:
  - Docker installed, Docker running, Linux containers;
  - cgroup version;
  - CPUs ≥ 2, memory ≥ 4 GiB;
  - inotify limits.

  Each result is `ok`, `warn` or `fail`, with a fix written for that platform. The machine sits behind a struct
  so tests pass a fake. `cka-sim doctor` runs it.
- **Review:** with Docker stopped, it prints `fail` with a fix and exits 1.
- **Learn:** where Docker's resource limits live on each OS.
- **Before:** `apps/lab/internal/doctor`.

## 3 · Accounts and a lab each

### 3.1 Accounts

- **Scope:**
  - `apps/auth`, its own app with Postgres and sql-migrate.
  - Salted password hashes, with a dummy compare for unknown emails.
  - A session JWT cookie plus a rotating refresh token. Reusing an old refresh token revokes the whole
    family.
  - The platform verifies the JWT with a shared secret and never touches the accounts database.
  - A login page.
- **Review:** a wrong password and an unknown email give the same answer; a refresh keeps you logged in.
- **Learn:** session JWTs versus refresh tokens, and rotation with reuse detection.
- **Before:** `apps/auth`, `apps/platform/internal/auth`, `apps/web/src/auth.ts`.

### 3.2 A lab per account

- **Scope:**
  - Each account gets its own lab (a pod per account on Kubernetes), touched set and exam.
  - An idle lab is stopped after `CKA_SIM_LAB_IDLE`.
  - The labs pool scales past one node.
- **Review:** two accounts each get their own lab on their own node; an idle one goes away.
- **Before:** `apps/platform/internal/desk`.

## 4 · The catalog and shipping

### 4.1 Grow the catalog

- **Scope:** the question bank (`docs/CKA-QUESTION-BANK.md`) and the `new-task` skill, then new tasks, each
  passing the selftest.
- **Before:** `.claude/skills/new-task/SKILL.md`.

### 4.2 CI

- **Scope:** `.github/workflows/ci.yml` runs `make test` and `make lint` on every push and pull request.
- **Review:** a failing test turns the commit red.
