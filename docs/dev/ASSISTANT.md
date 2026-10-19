# The assistant, in four parts

The goal: an AI tutor that knows the task you're on, can look at your cluster and run the grader, and helps you
learn instead of doing the task for you. It sits in the task page as a chat, beside the terminal.

The tools come first: they're what any AI needs to be useful here (our own chat agent, an AI CLI, an eval
harness), and you can try them from your own Claude Code before the page has a tutor.

Each part is one feature, reviewed and committed before the next, as in [BUILD-IN-STEPS.md](BUILD-IN-STEPS.md):
**Scope**, **Review** (what to try before accepting it) and **Learn** (the idea to understand first).

```
Part 1  the tools: TutorService on the platform ─► MCP server in apps/assistant   (your own AI CLI uses them)
Part 2  the tutor in the page: chat panel ─► platform ─► AssistantService         (a stand-in answers first)
Part 3  the AI: a Gemini loop on llm.Model over the MCP tools, in place of the stand-in
Part 4  evals: break a task, let the tutor diagnose it, score it against the grader
```

## The shape

```
browser ──POST /api/assistant/chat (SSE)──► platform ──Chat (Connect stream)──► apps/assistant   (Part 2)
                                               ▲                                     │ model loop (Part 3)
                                               └──── TutorService (Connect) ◄──── MCP tools        (Part 1)
```

- **Everything AI lives in `apps/assistant`:**
  - the MCP server and its instructions (`reviewer.md`);
  - the chat and, from Part 3, the model loop and its credentials;
  - the evals.

  It's its own Go module with the usual domain / use cases / adapters / api / cmd / e2e layout, and it imports
  nothing from the other apps.
- **The platform is the source of truth and the proxy.** It answers the tools through `TutorService`, and passes
  the page's chat to the assistant for the student it has identified, the way `/ws/terminal` passes the shell to
  the lab agent. The page never talks to the assistant directly.
- **Two contracts**, each a `.proto` in `apps/go-shared/contract/`, served and called with Connect:
  - `tutor/v1/tutor.proto` (Part 1), served by the platform: the read-only view of a student's task and cluster;
  - `assistant/v1/assistant.proto` (Part 2), served by the assistant: `Chat`, which takes the whole conversation
    and streams back the tools used and the answer. The assistant keeps no conversation.
- **Tokens:** `CKA_SIM_ASSISTANT_TOKEN` is shared by the platform and the assistant (`rpcauth`, as with the
  agent): the assistant sends it to `TutorService`, the platform sends it to `Chat`. The student is named in each
  call.

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
  - The student is fixed by config (`local` by default); the page's chat names its own (Part 2).
  - `make assistant` runs it (`CKA_SIM_ASSISTANT_ADDR`, `CKA_SIM_PLATFORM_URL`, `CKA_SIM_ASSISTANT_TOKEN` from
    `.env`); `claude mcp add --transport http cka-sim http://127.0.0.1:7373/mcp` connects it.
- **Review:** from your own Claude Code, on a task you set up:
  - "what am I working on?" names the task;
  - "review my attempt" runs the grader and explains the failing check without giving the answer;
  - "delete the pod" gets nowhere.
- **Learn:** MCP as the one seam between any AI and your system.

## Part 2 · The tutor in the page (a stand-in answers)

Build the whole path first with a harmless answer at the end of it, so the parts that can go wrong (streaming,
errors, the exam rule, the page) are reviewed before any model is involved.

- **Scope:**
  - `assistant.proto` with `AssistantService.Chat`: student, the task open in the page, the conversation; a
    stream of `tool` and `text` events back.
  - `apps/assistant`: a `chat` domain; `Reply` is a stand-in that reads the current task (a real tool event) and
    echoes the question. Served beside `/mcp` on the same port.
  - The platform: `POST /api/assistant/chat` (at most 40 messages, each at most 8 KiB, alternating and ending
    with the student), answered as server-sent events `tool`, `text`, then `done` or `error`. Refused during the
    student's exam (409); off without `CKA_SIM_ASSISTANT_URL` (503 `assistant_off`).
  - The page: **Ask the tutor** under the check result, with the tools it used, its answer as plain text, and
    errors that keep the question to ask again.
- **Review:** with `make dev` and `make assistant`, ask on a task page; set up the task and ask again, and the
  stand-in names it; stop the assistant and the panel says it doesn't answer.
- **Learn:** streaming an answer through two hops, and why the browser never holds the assistant's token.

## Part 3 · The AI

- **Scope:** `Reply` becomes a loop on `llm.Model` (Gemini on Vertex, `gemini-3.5-flash`) with `reviewer.md` as
  the system prompt and the MCP tools, streamed as it goes. A cap on tool calls and on time per answer. Without
  credentials, the assistant says it isn't configured.
- **Review:** "why does my check fail?" runs the check and points at the failing part without the answer; "show
  me the solution" uses `explain`; text in the cluster that reads like instructions doesn't steer it.
- **Learn:** a tool loop, and keeping the model behind an interface so another provider is one adapter.

## Part 4 · Evals

- **Scope:** for each task, set it up through the page's API, ask the tutor "my check fails, why?", and score
  whether it names what the failing checks test. Then run the solution and ask again: it should find nothing
  wrong. A second score: did a hint give the answer away?
- **Review:** a report per task and a total; a run on a handful of tasks is cheap enough to repeat.
- **Learn:** measuring an agent against ground truth you already have.

## Later

Ideas, each its own step:
- **An AI CLI in a terminal tab** (gemini-cli or Claude Code in a pty, as `/ws/terminal` does for the shell),
  locked down by managed settings: no shell, edit, write or web tools, only our MCP server.
- **Guards on tool output:** flag cluster text that reads like instructions to the AI.
- **The Kubernetes docs as a searchable reference (RAG)**, cited by link.
- **Learner notes** across tasks.
- **Markdown in the tutor's answers**, sanitized.
