# The assistant, in four parts

The goal: an AI tutor that knows the task you're on, can look at your cluster and run the grader, and helps you
learn instead of doing the task for you. In the end it runs in a second terminal next to the lab's, the way
Claude Code or gemini-cli runs in your own shell.

The tools come first: they're what any AI needs to be useful here (an AI CLI in the page, our own chat agent,
an eval harness), and you can try them from your own Claude Code before the page has an AI tab.

Each part is one feature, reviewed and committed before the next, as in [BUILD-IN-STEPS.md](BUILD-IN-STEPS.md):
**Scope**, **Review** (what to try before accepting it) and **Learn** (the idea to understand first).

```
Part 1  the tools: TutorService on the platform ─► MCP server in apps/assistant   (your own AI CLI uses them)
Part 2  evals: break a task, let an agent diagnose it, score it against the grader
Part 3  the AI tab: apps/assistant pty + platform proxy + xterm in the page        (a stand-in program first)
Part 4  the AI CLI in the pty, locked down, with credentials, pointed at the tools
```

## The shape

```
browser ──WebSocket /ws/assistant──► platform ──Connect stream──► apps/assistant ──pty──► AI CLI   (Parts 3, 4)
                                         ▲                               │
                                         └──── TutorService (Connect) ◄── MCP tools               (Part 1)
```

- **Everything AI lives in `apps/assistant`:**
  - the MCP server and its instructions (`reviewer.md`);
  - the evals;
  - the pty sessions, the CLI, its settings and its credentials.

  It's its own Go module with the usual domain / use cases / adapters / api / cmd / e2e layout, and it imports
  nothing from the other apps.
- **The platform is the source of truth and, later, a proxy.** It answers the tools through `TutorService` and
  bridges the page's WebSocket to the assistant, exactly as `/ws/terminal` bridges to the lab agent. The page
  never talks to the assistant directly.
- **Two contracts**, each a `.proto` in `apps/go-shared/contract/`, served and called with Connect:
  - `tutor/v1/tutor.proto` (Part 1), served by the platform: the read-only view of a student's task and cluster;
  - `assistant/v1/assistant.proto` (Part 3), served by the assistant: one bidirectional `OpenSession` stream that
    mirrors `OpenShell`.
- **Tokens:** the assistant calls `TutorService` with `CKA_SIM_ASSISTANT_TOKEN` (`rpcauth`, as with the agent),
  and names the student in each call. The token is only good for `TutorService`, which can't change anything.

## Part 1 · The tools

Built in two halves, each reviewed on its own.

### 1a · TutorService on the platform

- **Scope:**
  - `tutor.proto` with `TutorService`. Every call names a student; the platform refuses one it doesn't know
    (`NOT_FOUND`) and any call while that student's exam runs (`FAILED_PRECONDITION`):
    - `GetCurrentTask`: the task the student set up last and hasn't left (tidied). The platform remembers it per
      student, in memory;
    - `CheckTask`: the real grader; its PASS/FAIL lines are the only judge of a pass;
    - `InspectCluster`: one kubectl command, from an allowlist of read verbs (`get`, `describe`, `logs`,
      `events`, `top`, `explain`, `api-resources`, `auth can-i`) and flags. It runs as root on the control
      plane through the existing `RunScript`, every argument quoted, under `timeout 15`, its output cut to
      16 KiB. Anything else is `INVALID_ARGUMENT`;
    - `GetSolution`: the solution, for when the student asks for it.
  - Mounted at `/tutor.v1.TutorService/` beside the page's API, only when `CKA_SIM_ASSISTANT_TOKEN` is set.
  - Tests: unit tests for the allowlist (`delete`, `exec`, `--kubeconfig`, `-o go-template`, `;` and `$(…)`
    refused) and the current task; e2e over Connect against the fake agent.
- **Review:**
  - `InspectCluster` with `get pods -A` runs `timeout 15 kubectl 'get' 'pods' '-A'` on the control plane.
  - `delete pod x`, `get pods; rm -rf /` and `get pods --kubeconfig=/x` are refused before reaching the lab.
  - Set up a task in the page, and `GetCurrentTask` names it; leave the task, and it has none.
  - During an exam, every call is refused.
- **Learn:** an allowlist beats a denylist when a model builds the command; the grader, not the model, decides.

### 1b · The MCP server in apps/assistant

- **Scope:**
  - `apps/assistant` as its own module: `cmd/assistant` serving MCP (`github.com/modelcontextprotocol/go-sdk`,
    Streamable HTTP) on `127.0.0.1`, calling `TutorService` with the token.
  - Tools `current_task`, `run_check`, `inspect_cluster`, `explain` (the solution, only when asked), and
    `reviewer.md` (who it is: help you learn, don't do the task, stay on the CKA and this lab) as the server's
    instructions.
  - The student is fixed by config (`local` by default) until Part 3 gives each session its own.
  - `make assistant` runs it; `claude mcp add --transport http cka-sim http://127.0.0.1:7373/mcp` connects it.
- **Review:** from your own Claude Code, on a task you set up:
  - "what am I working on?" names the task;
  - "review my attempt" runs the grader and explains the failing check without giving the answer;
  - "delete the pod" gets nowhere.
- **Learn:** MCP as the one seam between any AI and your system.

## Part 2 · Evals

- **Scope:** for each task, run its setup, give an agent (our own loop on `llm.Model`, Gemini on Vertex) the
  tools and "my check fails, why?", and score whether it names what the check tests. Then run the solution and
  ask again: it should find nothing wrong. A second score: did a hint give the answer away?
- **Review:** a report per task and a total; a run on a handful of tasks is cheap enough to repeat.
- **Learn:** measuring an agent against ground truth you already have.

## Part 3 · The AI tab (no AI yet)

Build the whole pipe first, with a harmless program at the end of it, so the part that can go wrong (bytes,
resizes, ending a session) is reviewed before any AI is involved.

- **Scope:**
  - `assistant.proto` with `AssistantService.OpenSession`: the first message carries the terminal size and the
    student; then keys, resizes and output.
  - `apps/assistant`: a `sessions` domain (open, resize, close); a `pty` adapter that starts a command at the
    first size with `creack/pty`. The command is a stand-in, `cat`. Ending a session kills the process group.
  - The platform: `GET /ws/assistant`, bridging WebSocket ⇄ `OpenSession` the way the terminal bridge does; an
    assistant that can't be reached ends the socket with a reason the page shows.
  - The page: an **Assistant** tab beside the terminal, with its own xterm.
  - `make dev` and `make up` run it.
- **Review:** typing comes back echoed; resizing resizes the pty; closing the tab leaves no process behind;
  stopping the assistant shows "assistant unavailable".
- **Learn:** reusing the pty and stream pattern from the lab ([TERMINAL.md](../TERMINAL.md)).

## Part 4 · The AI CLI, locked down

Swap the stand-in for a real AI CLI behind a small `CLI` interface (its command, its settings files, its
environment), so a second one is one more adapter.

- **Which CLI:** gemini-cli on Vertex AI is recommended (GCP credit, default credentials, Workload Identity on
  GKE); Claude Code is the natural second adapter. Check the chosen CLI's current docs when building.
- **Scope:**
  - Each session gets a fresh working directory and `HOME`, runs as a non-root user, and ends when the CLI
    exits. There is no shell behind it.
  - Settings the user can't loosen: no shell, edit, write or web tools; no extensions; only our MCP server.
  - The context file (`GEMINI.md` or `CLAUDE.md`) holds `reviewer.md`; changing task offers a new session.
  - Credentials from the environment; with none, the session says the assistant isn't configured and ends.
  - One session per student, an idle timeout, a log line per session. Beginning an exam ends it.
- **Review:** "what is a PodDisruptionBudget?" gets an answer; asking it to run `ls`, write a file or fetch a
  URL is refused by its settings, not by its manners; during an exam the tab says it's off.
- **Learn:** managed settings are the security boundary, and the prompt is not.

## Later

Ideas from the first version (`archive/v2`, `IMPROVEMENTS-ASSISTANT.md`), each its own step:
- **Guards on tool output:** flag cluster text that reads like instructions to the AI.
- **The Kubernetes docs as a searchable reference (RAG)**, cited by link.
- **Learner notes** across tasks.
- **Our own chat pane on `llm.Model`**, without a CLI.
