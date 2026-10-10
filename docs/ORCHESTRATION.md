# Orchestration: how hosted cka-sim serves users

This is the target shape for running cka-sim as a website. [HOSTING.md](HOSTING.md) explains why we picked
Fly Machines. The big-picture flow is also drawn in EventCatalog (`docs/eventcatalog`, flow
`QuizAttemptLifecycle`).

## Two kinds of machines

```
                    ┌──────────────────────────── shared ────────────────────────────┐
Browser ──HTTPS/WSS──►  Main app (Go monolith, 2 Fly Machines)  ──►  Postgres         │
                    │        │                     │                                  │
                    └────────┼─────────────────────┼──────────────────────────────────┘
                             │ Fly Machines API    │ Fly private network (6PN)
                             ▼                     ▼
                    ┌──────── one per active quiz attempt ────────┐
                    │  Sandbox VM: agent + dockerd + kind cluster │
                    └─────────────────────────────────────────────┘
```

- **Main app.** One Go binary that every user shares. It runs as 2 Machines so a deploy or crash doesn't take the site down,
  and it's stateless because all state lives in Postgres. It also proxies the terminal websocket to the user's VM, so sandboxes are never
  exposed to the internet.
- **Sandbox VM.** One per active quiz attempt (not one per question; every task in the attempt shares the cluster, like the real
  exam). It's created when the user starts, ideally taken from a warm pool, and destroyed when the attempt ends.

## Domains inside the monolith

One binary and one deploy, with clear boundaries inside. A domain that ever needs to become its own service can move
out without a rewrite. The terminal is the most likely candidate.

| Domain | Owns | Exists today |
|---|---|---|
| identity | users, login, sessions | — |
| catalog | tasks (`task.md`, scripts), read-only from the embedded FS | `catalog/` |
| attempt | quiz/exam sessions: chosen tasks, timers, scores, Check/Submit | — |
| grading | runs `check.sh`, parses `PASS/FAIL <pts>` into a result | format in `catalog/lib.sh` |
| sandbox | VM lifecycle, warm pool, one-per-user rule, idle timeout, reaper | `internal/sandbox`: `Provider`, one `Lab`, the `Local` provider |
| terminal | browser websocket ⇄ shell | `internal/terminal`, `internal/server` |

The **agent** is a separate binary (`cmd/agent`) that runs inside each sandbox VM. It exposes `shell` and `run-script`
on the private network only, and it is mostly today's `DockerOpener` and script code moved into the VM.

### Boundary rules

- **Each domain owns its tables.** Other domains ask through Go calls and never query those tables directly.
- **Callers define small interfaces.** `terminal` defines `Opener`, `grading` defines `ScriptRunner`, and `sandbox` defines `Provider`.
- **Dependencies point one way.** `attempt → sandbox, catalog, grading`, `terminal → sandbox`, `sandbox → Provider`.
  Nothing calls back into `attempt`.
- **HTTP is a thin layer per domain.** Each domain registers its own routes and `cmd/server` only wires things together.
- **Local mode is just different wiring.** `Provider` = local kind, `Opener` = `DockerOpener`, and
  `ScriptRunner` = docker exec. The same domains work offline.

## Sandbox lifecycle

```
provisioning → ready → assigned → expired → destroying
```

- **Warm pool.** Keep N VMs in `ready` so "Start" is instant, and refill in the background.
- **Reaper.** A periodic job destroys `expired` VMs, plus any Fly Machine the database doesn't know about (leaks cost money).
- **No reuse between users.** A VM always goes to `destroying` after its attempt and never back to `ready`.

### One VM per user

A partial unique index enforces it in the database:

```sql
CREATE UNIQUE INDEX one_active_sandbox_per_user
  ON sandboxes (user_id)
  WHERE state IN ('provisioning', 'ready', 'assigned');
```

When "Start" finds an active sandbox, it either reconnects to it (page refresh, second tab) or refuses with
"you already have a running lab", offering to resume or end it.

### Timers

- **Attempt deadline**, e.g. 2h for an exam.
- **Idle timeout**, e.g. 15 min with no terminal connected, so abandoned tabs stop costing money.
- **Deploys drop websockets.** The shell lives in the VM and the agent keeps it alive for a few seconds,
  so the browser just reconnects.

## Main flow

1. **Start.** The user clicks Start quiz. `attempt` creates the attempt and asks `sandbox` for a VM: the user's existing one,
   one from the warm pool, or a new one via the Fly Machines API.
2. **Setup.** `attempt` → agent: run `setup.sh` for each task in the attempt.
3. **Work.** The browser gets the task list and opens the terminal websocket. `terminal` finds the user's VM through
   `sandbox` and proxies to the agent's shell, which runs `bash` in the kind control-plane node.
4. **Check / Submit.** `grading` → agent: run `check.sh`, parse the `PASS/FAIL <pts>` lines, and `attempt` stores the score.
   The VM never calls the main app, so it holds no credentials.
5. **End.** On submit, deadline, idle timeout or quit, `attempt` ends and `sandbox` marks the VM expired. The reaper then
   destroys it.

## Cost

The bill depends on **users with a lab open right now**, not total users. Each VM needs about 4 GB of RAM for kind.
50 users practising at the same time means 50 VMs, plus whatever the warm pool keeps idle.

## Rough repo layout (target)

```
apps/cka-sim/
  cmd/server    # wires all domains (hosted or local)
  cmd/agent     # runs inside each sandbox VM
  cmd/cli
  catalog/
  internal/
    identity/  attempt/  grading/
    sandbox/   (+ fly/, kind/ providers)
    terminal/  (+ docker opener, agent opener)
```

## Open questions

Not decided yet:

1. **How domains react to each other.** When an attempt ends, should `attempt` call `sandbox.Release()` directly
   (simple, explicit), or publish an in-process "attempt ended" event that `sandbox` listens to (more decoupled, harder to follow)?
2. **Is grading its own domain or part of attempt?** It's small, but the CLI and the lab "Check" button use it outside
   exams too.
3. **Database separation.** Use one Postgres schema per domain (the database enforces the boundary) or shared tables with a convention?
4. **Resolving "this user's VM".** Should `terminal` ask `sandbox` on every connect, or should `attempt` issue a signed ticket that
   carries the VM address?
5. **How much to trust scores.** The user is root in the cluster and could tamper with checks. That's fine for practice, but leaderboards or
   certificates would need checks that run outside the user's reach.
