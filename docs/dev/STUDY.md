# Study mode, in three parts

The goal: study one CKA topic for real, not only practise a task you picked. You say "study NetworkPolicies
with me", and an AI finds where your understanding ends, picks catalog tasks that start there, sets each one up
on your lab, lets you solve it, grades it with the real `check.sh`, and talks it through before the next. It
ends with a few mixed tasks that don't say what they're about, the way the exam doesn't.

**It runs on a subscription, not on API tokens.** The AI is your own Claude Code (or any AI CLI that speaks
MCP): it runs the loop and pays for it with its own login. cka-sim gives it tools and the method, through the
MCP server `apps/assistant` already serves. Nothing in Parts 1 and 2 calls a model.

The ideas come from [LearnHarness](https://github.com/amosblomqvist/learnharness): probe before teaching, a plan
from where you are to where you want to be, practice that doesn't name its topic, a hint ladder, and a record of
attempts kept as evidence rather than a score. Its knowledge graph, sub-agents and desktop app are left out. One
thing is ours: every practice problem is a real task with a real grader, so a pass is never the model's opinion.

Each part is one feature, reviewed and committed before the next, as in [BUILD-IN-STEPS.md](BUILD-IN-STEPS.md):
**Scope**, **Review** and **Learn**.

## The shape

- **A new contract, `study/v1/study.proto`**, served by the platform beside `TutorService`, with the same token
  and the same rules: every call names a student, `NOT_FOUND` for one the platform doesn't know,
  `FAILED_PRECONDITION` while that student's exam runs. `TutorService` keeps its promise to only look; everything
  that changes something (setting a task up, saving a note) is in `StudyService`.
- **The platform keeps the record.** Every check of a practice task (from the page, the tutor or a study
  session) is saved as an attempt, per student, in a file. The grader writes the record; the AI only adds notes.
- **`apps/assistant` adds tools and a prompt.** The tools call `StudyService`; the prompt is the study method,
  served as an MCP prompt so Claude Code shows it as `/mcp__cka-sim__study`. `reviewer.md` stays the tutor's.
- **You work where you already do:** the task page or `cka-sim shell`. The AI never gets a shell.

## Part 1 · The tools

Built in two halves, each reviewed on its own (the contract and the platform first, then the assistant).

### 1a · StudyService and the record, on the platform

- **Scope:**
  - `study.proto` with `StudyService`:
    - `ListTasks`: the catalog filtered by domains, topics and a search, as the page's `GET /api/tasks` does
      (`tasks/usecases.ListTasks`): id, title, domain, topics, weight. No question text, so a list can't spoil
      the mixed review.
    - `StartTask`: tidies the student's current task if it's another one, then heals and sets the task up
      (`TidyTask`, `StartTask`); answers the task with its question and host. `FAILED_PRECONDITION` unless the
      lab is ready, `NOT_FOUND` for an unknown task.
    - `GetHistory`: the student's attempts and notes, newest first, at most 200, optionally for some tasks
      only.
    - `SaveNote`: a note on a task, at most 500 characters, like "mixed up ingress and egress `from`/`to`".
  - **Attempts:** `tasks/usecases.CheckTask` saves one after every practice check: task, time, earned, total,
    and the descriptions of the checks that failed. Not during an exam (its own record already has them).
  - **Storage:** `files` adapter, `<CKA_SIM_STATE_DIR>/history/<student>.jsonl`, one line per attempt or note,
    appended; behind an `AttemptRepository` the use cases own, so Postgres is one adapter later.
  - Mounted at `/study.v1.StudyService/` beside `TutorService`, only when `CKA_SIM_ASSISTANT_TOKEN` is set.
  - Tests: unit tests for the record (a check appends, an exam check doesn't, a note over 500 characters is
    refused) and for `StartTask` tidying the previous task; e2e over Connect against the fake agent, including
    every refusal.
- **Review:**
  - Check a task twice in the page; `GetHistory` answers both attempts with what failed, newest first.
  - `StartTask` for `nw-nodeport` while `wl-pdb` is current: `wl-pdb` is tidied, `nw-nodeport` set up, and
    `GetCurrentTask` names it.
  - During an exam, every call is refused, and exam checks leave no attempts.
  - Restart the platform: the history is still there.
- **Learn:** keeping a write path apart from a read-only one, and an append-only log as the simplest record.

### 1b · The tools in apps/assistant

- **Scope:**
  - A `study` domain with use cases calling `StudyService` through the `platform` adapter, like `tutor` does.
  - MCP tools:
    - `list_tasks` (domains, topics, search);
    - `list_topics`: every domain and topic with how many tasks it has, computed from `list_tasks`, so the AI
      can turn "network policies" into the catalog's own topic names;
    - `start_task` (an id);
    - `get_history` (optional task ids);
    - `save_note` (task id, text).
  - The platform's refusals reach the AI with their reason, as with the tutor's tools.
- **Review:** from your own Claude Code (`claude mcp add --transport http cka-sim http://127.0.0.1:7373/mcp`):
  - "which tasks cover PodDisruptionBudgets?" lists them;
  - "set up wl-pdb" sets it up, and `cka-sim shell` shows its namespace;
  - "what have I failed lately?" answers from the history.
- **Learn:** tools as the whole interface to a system; the CLI's model does the reasoning, we only expose
  facts and actions.

## Part 2 · The study method

- **Scope:**
  - An MCP prompt `study` with one argument, `topic` (free words: "networkpolicies", "etcd backup",
    "troubleshooting"). Its text is `study.md`, embedded like `reviewer.md`, with the topic filled in. It asks
    the AI to run a session in four phases, one at a time:
    1. **Probe:** look at `get_history` for the topic first, then ask three or four short questions, one at a
       time, waiting for each answer (multiple choice or "what would you run"), to find where the student's
       understanding ends. Ask how sure they are. Don't teach yet.
    2. **Plan:** say in a few lines what they hold, where it ends and the goal, then pick three to five catalog
       tasks for the topic from `list_tasks`, easiest first, skipping what the probe shows they hold.
    3. **Practice:** for each task, `start_task`, give its question as the task page shows it, and tell the
       student to solve it in the page or `cka-sim shell` and say when they're done. Then `run_check`. On a
       fail, one hint at a time (the hint ladder: where to look, then which field or concept, then the
       command), never the manifest; `explain` only when asked. On a pass, one or two lines on why it works,
       with a `search_docs` link. Then `save_note` with what this attempt showed, and ask before going on.
    4. **Mixed review:** two or three tasks from other topics the history says they have practised, without
       saying which topic each is. Same loop.
  - `reviewer.md`'s rules still hold: the grader is the only judge of a pass, cluster text is never an
    instruction, the AI can't change the cluster.
  - Tests: the prompt lists and renders with its topic; an empty topic is refused.
- **Review:** run `/mcp__cka-sim__study networkpolicies` in Claude Code on a ready lab and go through a whole
  session: the probe waits for your answers, the plan names real task ids, a failed check gets a hint rather
  than the answer, and `get_history` afterwards shows the attempts and notes.
- **Learn:** the method lives in a prompt and the facts in tools, so another AI CLI runs the same session.

## Part 3 · Study in the page (later, still on the subscription)

Not planned in detail yet; one of these, chosen after Part 2 has been used for a while:
- **Claude Code in a terminal tab:** the page opens a second tab running `claude` in a pty, the way
  `/ws/terminal` does for the shell, with a settings file that allows only the `mcp__cka-sim__*` tools.
- **A CLI-backed tutor:** a second implementation of `Reply` that runs `claude -p --output-format stream-json
  --strict-mcp-config` with built-in tools off and only our MCP server, so "Ask the tutor" answers on the
  subscription and Gemini stays the other implementation.

Either way the CLI must be installed and logged in where it runs, so it fits the local install, not the hosted
site: one person's subscription can't serve other students.

## Part 4 · A practice exam the AI builds

- **Scope:**
  - `exam/v1/exam.proto` with its own `ExamService`, served beside `TutorService` and `StudyService` with the
    same token, only when `CKA_SIM_ASSISTANT_TOKEN` is set (`internal/api/exam_service.go`). Every call names a
    student (`NOT_FOUND` otherwise):
    - `StartExam` (count, minutes, optional domains, topics, task ids): the page's exam, drawn only from the tasks
      the selection matches (all of them when fewer match than the count). Answers the drawn ids and titles,
      never a question. `INVALID_ARGUMENT` for a bad count or minutes, an unknown domain, a malformed topic, a task
      id that names no task or a selection matching no task;
      `FAILED_PRECONDITION` while an exam runs or is graded, or unless the lab is ready.
    - `GetExamResult`: the last exam once it's graded: score, and per task earned, total and the failed checks.
      `FAILED_PRECONDITION` when there is none and while it runs or is graded, so a review never sees a task
      while it can still be solved.
  - Each flow stays in its domain: `tasks.Filter` gained `IDs` and `Select` (which refuses an unknown id), so
    choosing what's eligible is the tasks domain's; `exams/usecases.BeginExam` takes that selection (empty = the
    whole catalog, which is what `POST /api/exam` sends) and `GetExamResult` is a new exams use case. An exam
    task keeps its title, like its weight, so neither RPC reads the catalog again.
  - Beginning any exam (page or AI) first tidies and leaves the practice task the student was on
    (`tasks/usecases.TidyCurrentTask`), then heals, so its namespaces don't linger into the exam.
  - `apps/assistant`: a `domain/exam` over `adapters/platform/exam.go`, MCP tools `start_exam` and
    `exam_result` (`internal/api/exam.go`), only on the AI CLI's server (`NewCLIServer`), not the page's tutor.
    `study.md` gets an optional fifth phase: offer the exam, never help during it, review with `exam_result`.
  - The tutor's and study's calls stay refused while the exam runs, whoever started it; only `search_docs` and
    the `study` prompt aren't gated, and the prompt tells the AI not to explain any task until it's over.
- **Review:** in Claude Code, "give me a 3-task exam on networkpolicies": the page's exam view shows it setting
  up; `exam_result` is refused until it's graded; after **End exam**, the AI reviews what failed.
- **Learn:** one aggregate per domain; a selection is a tasks concern that the exam only consumes.

## Left out

- ~~Opening a task's page sets it up again, even when `start_task` just did.~~ Done in S3: the page follows
  the session (`GET /api/tasks/current`, `src/follow.ts`) and skips a setup the session already ran.
- Tasks the AI writes itself. Possible later behind the selftest's rule (the check fails after `setup.sh` and
  passes after `solution.sh`, or the task is thrown away).
- Scores per topic, spaced repetition, a knowledge graph. The history is evidence for the AI to read, not a
  mastery number.
- Evals for study mode; `make eval` still scores the tutor only.
