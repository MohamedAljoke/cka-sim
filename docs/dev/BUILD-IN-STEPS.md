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

## 3 · Students and a lab each

### 3.1 Students by signed token

No database and no login page. A student is a token signed with the server's key, so the token itself is the
account.

> [!NOTE]
> A first, simpler version is built (see [STATUS.md](STATUS.md)): a fixed `CKA_SIM_STUDENTS` list of
> `name=token`, joined at `/api/join?token=`, with an empty list meaning everyone is `local`. What's left is
> the signed token below: `mint`, expiry and the denylist, so students change without a restart.

- **Scope:**
  - **The token:** an HS256 JWT `{sub, exp}` signed with `CKA_SIM_SECRET`.
  - **The verifier**, a small piece in the platform:
    - it reads the `cka_sim_student` HttpOnly cookie;
    - it checks the signature, `exp` and a denylist (`CKA_SIM_REVOKED`, a list of `sub`);
    - it puts `sub` in the request context;
    - with no valid token it answers 401 `student_unknown`, and the page says to ask for an invite link.
  - **Invite links:**
    - `GET /?key=<token>` sets the cookie and redirects, so the token leaves the URL;
    - `server mint -name alice -days 30` prints the link;
    - `make -C infra/gke student NAME=alice` runs it inside the app pod, so the secret never leaves the
      cluster.
  - **Locally:**
    - the secret is generated once into the state directory, like the agent token;
    - `CKA_SIM_STUDENTS=auto` gives `sub: local` to any request without a cookie, so `cka-sim up` and
      `make dev` behave exactly as today;
    - hosted uses `invite`.
  - To everything else the owner is just a string. A login page added later would be one more way to get the
    same cookie.
- **Review:**
  - Locally, nothing changes.
  - Hosted:
    - with no link, the page asks for one;
    - an invite link gets you in;
    - an edited token is refused;
    - putting its `sub` on the denylist ends it.
- **Learn:** signed tokens as stateless identity (what a signature proves and what it doesn't), HttpOnly
  cookies, and revocation without a database.

### 3.2 A lab per student

> [!NOTE]
> Mostly built: the fleet, per-owner use cases and exams, the `max_labs` cap, `make labs`. The pod is named
> `lab-<student>` rather than a hash, since a student's name is already DNS-safe. Left: the idle stop, the
> "site is full" message, and the live review below.

- **Scope:**
  - **`kube.Pod`:** named `lab-<short hash of sub>` (pod names must be DNS-safe), and labelled
    `cka-sim/owner=<hash>`. `Up`, `Down`, `Where` and the pod spec don't change.
  - **A registry, `Labs.For(owner) *labagent.Lab`:** it makes each lab lazily. Locally, every owner gets the
    one `Fixed` lab.
  - **The use cases take the owner:**
    - `LabProvider` and `ScriptRunner` get an owner argument;
    - the touched set and the script lock (`cluster`) are one per owner;
    - the exam is `<state>/exams/<owner>.json`, with a session per owner;
    - handlers and `/ws/terminal` pass the `sub` from the request.
  - **Idle stop:** a lab with no request for `CKA_SIM_LAB_IDLE` (default 1h) is stopped. The Role gains `list` on
    pods.
  - **The cap:** the labs pool's `max_node_count` becomes `var.max_labs`, the cap on concurrent labs. Past it,
    the pod waits for a node, and the page says the site is full.
  - **Seeing the fleet:** `make -C infra/gke labs` lists each lab pod with its owner, phase, node and age, and the
    labs pool's nodes.
- **Review, live on GKE:**
  - **The fleet grows.** Two invite links in two browser profiles start two labs. `make labs` shows two pods on
    two nodes. Stopping one removes only that pod, and its node a few minutes later.
  - **Students can't reach each other.** From student A's terminal, `curl -m 5 <B's pod IP>:7171` and
    `ssh <B's pod IP>` time out:
    - the NetworkPolicy lets only the app into a lab, and only on 7171;
    - labs have no egress to `10.0.0.0/8`, which holds every pod IP.
  - **The work stays separate.** A's practice and exam never show for B.
  - **Idle labs stop.** A lab left idle past the limit stops by itself.
- **Learn:** the owner as a parameter rather than objects per user; a fleet whose size is a node pool's
  maximum; testing isolation instead of assuming it.

## 4 · The catalog and shipping

### 4.1 Grow the catalog

- **Scope:** the question bank (`docs/CKA-QUESTION-BANK.md`) and the `new-task` skill, then new tasks, each
  passing the selftest.
- **Before:** `.claude/skills/new-task/SKILL.md`.

### 4.2 CI

- **Scope:** `.github/workflows/ci.yml` runs `make test` and `make lint` on every push and pull request.
- **Review:** a failing test turns the commit red.
