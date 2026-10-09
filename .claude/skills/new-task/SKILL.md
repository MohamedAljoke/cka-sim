---
name: new-task
description: Turn a Kubernetes/CKA concept into a cka-sim practice task in apps/cka-sim/catalog/<id>/ (task.md, setup.sh, check.sh, solution.sh, explain.md), then prove it on the running cluster. Use when the user says "make this a task", "turn X into a task", "add a task for <concept>", or picks a question from docs/CKA-QUESTION-BANK.md.
---

# New cka-sim task

The user is studying for the CKA. They found a concept and want to practise it as an exam-style
task. Work in pair mode: show the **idea first**, then write the files once they agree.

## 1. Pin down the task (before writing any file)

1. Find the concept in `docs/CKA-QUESTION-BANK.md`. Reuse its ID area (T/A/N/W/S) and check its
   **Needs** column. If it's not there, decide which domain it belongs to and add it to the bank later.
2. Check the cluster can host it. Today it's kind with `cka-sim-control-plane`, `cka-sim-worker` and
   `cka-sim-worker2` (`internal/cluster/cluster.go`). There is no metrics-server, no
   policy-enforcing CNI, no Gateway API and no Helm. If the task needs one of these, say so and stop:
   the cluster comes first.
3. Look at `ls apps/cka-sim/catalog/` so the new task doesn't duplicate an existing one.
4. Propose the task in a few lines and wait for an OK:
   - id, title, domain and host
   - the scenario: what setup breaks or leaves undone
   - what the user must do
   - each check, with its points

## 2. Write the folder

Copy the shape of `apps/cka-sim/catalog/tr-service/` (troubleshooting) or `wl-scale/` (build
something). Read both before you start.

### `task.md`

```markdown
---
id: <domain-prefix>-<topic>        # must equal the folder name
title: <short, says the goal>
domain: troubleshooting | architecture | networking | workloads | storage
host: cka-sim-control-plane        # the node the user ssh-es into
topics: [kebab-case, words]
weight: <1-8, rough exam points>
---
## Context        (optional: what exists, in exam voice)

## Task

<numbered steps, exact names, namespaces, ports and file paths in `code`>
```

- **Id prefixes:** `tr-` troubleshooting, `ar-` architecture, `nw-` networking, `wl-` workloads,
  `st-` storage.
- **No other frontmatter fields.** The parser is strict and the server won't start with unknown
  fields such as v1's `cluster:` or `order:`.
- **Exam voice:** state the goal, never the fix. Add constraints like "Do not modify Deployment X"
  when they matter.

### `setup.sh`

- It runs **inside the host node**, where `kubectl` is admin, after `catalog/lib.sh`. `TASK_ID` is set.
- It must be **idempotent**: opening the task again resets it. Start with `fresh_ns <namespace>`,
  and give every task its own namespace.
- It builds the scenario and **waits until it is in place** (`rollout status`, `wait_for`).
- It can't hide state: anything it writes on the node, the user can read. For "don't modify X"
  checks, prefer facts you can derive, such as `metadata.generation` being `1` on a freshly
  created object, over stored values.

### `check.sh`

- Every point is one line: `check <points> "<what it proves>" <command...>`.
- Helpers in `lib.sh`: `eq`, `wait_for`, `fresh_ns`.
- Check **the result, not the method**. `wl-scale` checks Ready Pods, not that `kubectl scale` ran.
- After setup, **every check must fail**. If a check could pass with no work done, gate it, as
  `wl-scale` does with its `if ... else echo "FAIL ..."`.
- Give eventual-consistency things time: `wait_for 20 <fn>` before checking.
- Put helper functions at the top, then the `check` lines.

### `solution.sh`

- It is the shortest exam-speed fix, using imperative `kubectl` where possible.
- Comment lines can show the diagnosis commands a candidate would run.

### `explain.md`

- Explain **why** the fix works: which controller or component does what. Name the trap.
- Then show the commands in a `sh` block, the faster exam alternative, and how to verify.

## 3. Prove it

```sh
cd apps/cka-sim
go test ./catalog/...          # the frontmatter parses, the host is a real node, no script is empty
make try TASK=<id>             # setup → every check FAILs → solution → every check PASSes
```

`make try` needs the cluster up (`make up`). Also break the "don't modify" rule by hand once and
re-run `check.sh` to see that check fail. The command is in `scripts/try-task.sh`.

If the task came from the question bank, fill in its **Covered** cell with `main \`<id>\``.

## Rules for this repo

- Don't commit; the user commits.
- Only add comments for a non-obvious *why*.
- Restart `make dev` to see the task in the web list, because the catalog is embedded at build time.
