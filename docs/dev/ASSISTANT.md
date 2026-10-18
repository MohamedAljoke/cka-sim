# The assistant, in three parts

The goal: next to the lab's terminal, the page opens a second terminal where an AI CLI runs. You talk to it
the way you would to Claude Code or gemini-cli in your own shell. It knows the task you're on, can look at the
cluster and run the grader, and helps you learn instead of doing the task for you.

Each part is one feature, reviewed and committed before the next, as in [BUILD-IN-STEPS.md](BUILD-IN-STEPS.md):
**Scope**, **Review** (what to try before accepting it) and **Learn** (the idea to understand first).

```
Part 1  apps/assistant + contract ─► platform proxy ─► AI tab in the page   (a stand-in program, no AI yet)
Part 2  the AI CLI in the pty, locked down, with credentials                 (talk to an AI)
Part 3  task context: instructions + MCP tools on the platform              (it knows your task)
```

## The shape

```
browser ──WebSocket /ws/assistant──► platform ──Connect stream──► apps/assistant ──pty──► AI CLI
                                         ▲                               │
                                         └──── read-only task API ◄── MCP tools (Part 3)
```

- **Everything AI lives in `apps/assistant`:**
  - the pty sessions;
  - the CLI, its settings and its credentials;
  - the instructions;
  - the MCP server.

  It's its own Go module with the usual domain / use cases / adapters / api / cmd / e2e layout, and it imports
  nothing from the other apps.
- **The platform is only a proxy and the source of truth.** It bridges the page's WebSocket to the assistant,
  exactly as `/ws/terminal` bridges to the lab agent, and it answers the assistant's tool calls about tasks and
  the cluster. The page never talks to the assistant directly.
- **The contract** is a `.proto` in `apps/go-shared/contract/assistant/v1/`, served and called with Connect, like
  `lab.proto`. Its one call is a bidirectional `OpenSession` stream that mirrors `OpenShell`: the first message
  carries the terminal size (and, from Part 3, the task), then keys, resizes and output.
- **Tokens, with no accounts yet:**
  - The platform calls the assistant with a bearer token (`rpcauth`, as with the agent).
  - The assistant calls the platform back with a second token, which is accepted only on the read-only
    assistant routes.

## Part 1 · apps/assistant and the proxy (no AI yet)

Build the whole pipe first, with a harmless program at the end of it, so the part that can go wrong (bytes,
resizes, ending a session) is reviewed before any AI is involved.

- **Scope:**
  - `assistant.proto` with `AssistantService.OpenSession`. `make proto` generates it.
  - `apps/assistant`:
    - `cmd/assistant` on `:7373`, with `/healthz` and the token;
    - a `sessions` domain (open, resize, close);
    - a `pty` adapter that starts a command at the first size with `creack/pty`.
    - The command is a stand-in, `cat`, which echoes what you type. Ending a session kills the process group.
  - The platform:
    - `GET /ws/assistant`, which bridges WebSocket ⇄ `OpenSession` the way the terminal bridge does;
    - `CKA_SIM_ASSISTANT_URL` and `CKA_SIM_ASSISTANT_TOKEN`;
    - an assistant that can't be reached ends the socket with a reason the page shows.
  - The page: an **Assistant** tab beside the terminal, with its own xterm.
  - `make dev` and `make up` run it: `infra/assistant/assistant.Dockerfile` and a service in
    `infra/local/compose.yml`.
  - Tests:
    - unit tests for the session and the bridge;
    - e2e on the assistant: open, type, resize, close;
    - e2e on the platform: the socket reaches a fake assistant.
- **Review:**
  - Typing in the Assistant tab comes back echoed.
  - Resizing the panel resizes the pty, which the e2e checks.
  - Closing the tab leaves no process behind.
  - Stopping the assistant shows "assistant unavailable".
- **Learn:** one app per responsibility, behind a contract; reusing the pty and stream pattern from the lab
  ([TERMINAL.md](../TERMINAL.md)).

## Part 2 · The AI CLI, locked down

Swap the stand-in for a real AI CLI. The CLI sits behind a small `CLI` interface in the assistant (its command,
its settings files, its environment), so a second one is one more adapter.

- **Which CLI** is the first decision of this part. **gemini-cli on Vertex AI** is recommended:
  - It's paid from the GCP credit, and it authenticates with Google's default credentials, the same way the
    old `llm/vertex` did.
  - On GKE that becomes Workload Identity, with no key file.

  Claude Code is the natural second adapter. Check the chosen CLI's current docs (settings file locations,
  MCP config, tool allow and deny lists) when building, not from memory.
- **Scope:**
  - The assistant image installs the CLI.
  - Each session gets a fresh, private working directory and `HOME`, runs as a non-root user, and ends when
    the CLI exits. There is **no shell** behind it.
  - Settings the user can't loosen (the CLI's system-wide settings file):
    - no shell, edit, write or web tools;
    - no extensions;
    - only the MCP servers we configure (none until Part 3).
  - Credentials from the environment (`GOOGLE_CLOUD_PROJECT`, `GOOGLE_CLOUD_LOCATION`, ADC mounted
    read-only locally). With none, the session prints one line saying the assistant isn't configured and
    ends.
  - Limits:
    - one session per lab;
    - an idle timeout (`CKA_SIM_ASSISTANT_IDLE`);
    - a log line per session (start, end, duration).
- **Review:**
  - "what is a PodDisruptionBudget?" gets an answer in the tab.
  - Asking it to run `ls`, write a file or fetch a URL is refused by its settings, not by its manners.
  - Closing the tab ends the CLI.
- **Learn:**
  - A CLI agent is a program in a pty like any other.
  - Managed settings are the security boundary, and the prompt is not.
  - Credentials stay in the one app that needs them.

## Part 3 · Context on the task

Make it useful for this lab: it should know the task, read the cluster and run the grader. It should never be
able to change anything.

- **Scope:**
  - **The task in the session.**
    - The page sends the open task's id with `OpenSession`.
    - The assistant writes the CLI's context file (`GEMINI.md`, or `CLAUDE.md`) into the session's
      directory: `reviewer.md` (who it is, to help you learn and not do the task, to stay on the CKA and this
      lab) plus the task's id and question.
    - Changing task offers a new session.
  - **MCP tools.**
    - The assistant serves an MCP server (`github.com/modelcontextprotocol/go-sdk`, Streamable HTTP) on
      `127.0.0.1` only. The CLI's settings point at it.
    - Every tool is read-only and calls the platform with the assistant's token:
      - `current_task`: the question and the lab state;
      - `run_check`: the real grader; its PASS/FAIL lines are the only judge of a pass;
      - `inspect_cluster`: kubectl with only read verbs (`get`, `describe`, `logs`, `top`, `events`) and an
        allowlist of flags, run through the existing `RunScript` with `timeout 15` and capped output;
      - `explain`: the solution, only when you ask for it.
  - **The platform's side.**
    - `/api/assistant/*` routes accept only the assistant token. They are read-only, and refuse during an
      exam (409 `exam_running`).
    - Beginning an exam ends any open assistant session.
- **Review:**
  - "what am I working on?" names the task.
  - "review my attempt" runs `run_check` and explains the failing check without giving the answer.
  - "delete the pod" gets nowhere: there is no tool for it, and `inspect_cluster` refuses the verb.
  - During an exam, the tab says it's off.
- **Learn:**
  - MCP as the one seam between any AI and your system.
  - An allowlist beats a denylist when a model builds the command.
  - The grader, not the model, decides pass or fail.

## Later

Ideas from the first version (`archive/v2`, `IMPROVEMENTS-ASSISTANT.md`), each its own step once the three parts
hold:
- **Guards on tool output:** flag cluster text that reads like instructions to the AI.
- **An eval** that plays solved and untouched attempts and scores the reviews against the grader.
- **The Kubernetes docs as a searchable reference (RAG)**, cited by link.
- **Learner notes** across tasks.
- **Our own chat agent on `llm.Model`**, for a chat pane without a CLI.
