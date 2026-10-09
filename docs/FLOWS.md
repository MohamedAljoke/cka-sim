# Flows

How cka-sim works end to end, told as what happens when you do something. Each flow follows one action
from the command or click to the container it reaches, naming the file, endpoint or command at every hop.
New flows go here in the same shape: the trigger, then each hop in order, then what you see.

## 1. `make dev`: from the command to a ready server

`make dev` first runs `make up`, which is `cka-sim up` (`cmd/cli`). That creates a kind cluster of three
containers built from `internal/cluster/node.Dockerfile`: `cka-sim-control-plane`, `cka-sim-worker` and
`cka-sim-worker2`. Each node has a `candidate` user with sudo, sshd, and `k` with completion.

`up` then calls `Cluster.Prepare` (`internal/cluster/base.go`). It builds `base.Dockerfile` and starts
`cka-sim-base` on the `kind` network, so the node names resolve from it. Base has no kubectl. `Prepare` gives
base's `candidate` an ssh key and, on every node, installs that key in `authorized_keys` and the cluster's
admin kubeconfig for both `candidate` and `root`. If base already exists, `Prepare` does nothing, so open
terminals survive a second `up`.

Next, `make dev` builds `bin/server` and starts it in the background, then starts Vite on 5173.
`vite.config.ts` forwards `/api` and `/ws` to the Go server on `127.0.0.1:7070`. The server
(`cmd/server/main.go`) claims the port and checks that the cluster and base are up (`cluster.RequireUp`). It
loads the catalog (`tasks.Load(catalog.FS)`: every `catalog/<id>/task.md` and its scripts, built into the
binary) and `lib.sh`. It closes any shells left over from a previous run (`EndAll`), then serves.

Ctrl-C reaches the server directly, because it runs as a binary and not through `go run`. The Makefile's
trap waits for it, and on the way out the server closes every page shell.

## 2. Opening the page: the terminal

`main.ts` opens a websocket to `/ws/terminal`. The server's `DockerOpener` (`internal/terminal/docker.go`)
answers with the equivalent of `docker exec -it -u candidate -w /home/candidate cka-sim-base bash` and pipes
bytes both ways between xterm and that shell. So you land on `candidate@base` with no kubectl, as in the
exam. When the tab closes, the shell is hung up as the same user. Root can't do it, because reading another
user's process environment needs ptrace, which Docker withholds.

At the same time, `showTasks` (`src/tasks.ts`) calls `GET /api/tasks` and draws the numbered list. The
server sends only metadata (title, domain, topics, weight, host), never the scripts.

## 3. Clicking a task: question and setup

`openTask` sends two requests in parallel:

- `GET /api/tasks/{id}/question` returns the question markdown, and `marked` renders it right away.
- `POST /api/tasks/{id}/start` reaches `startTask` (`internal/server/api.go`). It takes the setup lock; if a
  setup is already running, it answers 409. It then calls `tasks.Start` (`internal/tasks/run.go`), which
  reads `setup.sh` and hands it to the runner with a 3-minute limit.

The runner (`internal/runner/runner.go`) runs `docker exec -i -e TASK_ID=<id> <task's host> bash -s` and
writes `lib.sh` followed by `setup.sh` to its stdin. The script runs as root on the node the task names,
using root's kubeconfig. Helpers such as `fresh_ns` wipe and recreate the task's namespace before the
script builds the broken state. The request keeps going even if the page goes away (`WithoutCancel`), so a
setup is never left half done.

The page shows "Preparing the cluster…", then "Ready" on 204, or the script's output if setup failed. You
then follow the "Connect first" line: `ssh <host>` logs you in with base's key and no password, and you
solve the task with `k`.

## 4. Check, Solution, Reset

**Check.** You solve the task in the page terminal (`ssh <host>`, then `k …`) and press **Check**. The page
disables Check and Reset, shows "Checking…" and sends `POST /api/tasks/{id}/check`. `checkTask`
(`internal/server/api.go`) takes the same lock as setup, so a check during setup gets 409 "a script is
already running". It calls `tasks.Check` (`internal/tasks/run.go`), which runs `check.sh` on the task's
host through the same runner as setup, with `lib.sh` in front. Each `check <points> <description> <cmd>`
in the script prints `PASS` or `FAIL` with its points and description. `grader.Parse`
(`internal/grader/grader.go`) keeps only those lines, ignoring kubectl noise, and adds up earned and total.
The page shows a bold `2 / 4` and one ✓/✗ line per check. If `check.sh` itself crashes, the server answers
500 with its output, and the page shows "Check failed: …" rather than a silent 0.

**Solution.** **Solution** sends `GET /api/tasks/{id}/solution`, which returns the task's `explain.md`,
loaded at startup. No script runs. The page renders it with `marked` under the question; pressing again
hides it.

**Reset.** **Reset** clears the result, sets the status back to "Preparing the cluster…" and sends the same
`POST /api/tasks/{id}/start` as opening the task. `fresh_ns` wipes the namespace and `setup.sh` rebuilds
the broken state. While it runs, Check is disabled.

## 5. `make selftest`: proving every task is fair

`make selftest` runs `go test ./catalog -run TestSelftest -selftest`. The `-selftest` flag keeps
`go test ./...` from touching the cluster. `catalog/selftest_test.go` loads every task and `lib.sh` and
builds the same `runner.Node` the server uses. For each task, `tasks.Selftest` runs setup, then check,
which must earn 0 of more than 0 points (nothing to earn for doing nothing). It then runs `solution.sh`
(`tasks.Solve`) and checks again, which must earn every point. A failure names the task and the checks
at fault, for example:
`wl-scale: earns 0/4 after the solution: "Deployment web wants 4 replicas", …`.

## 6. Exam mode

**Starting.** Above the practice list, **Start exam** opens a small form (`apps/web/src/tasks.ts`): how many
tasks (1 to the catalog size, default 16 or fewer) and the minutes (default 120). Submitting sends
`POST /api/exam {count, minutes}`. `beginExam` (`internal/server/exam.go`) checks the numbers (400 if they are
out of range) and takes the script lock, so a running practice setup gets 409. It then calls
`exam.Session.Begin` (`internal/exam/session.go`). `tasks.Draw` (`internal/tasks/draw.go`) picks the tasks:
each domain gets its share of the count by the CKA weights, with any shortfall moved to the heaviest domains,
and the result is shuffled. The exam is saved with every task `preparing` to
`~/.local/state/cka-sim/exam.json`, and the request answers 202 right away. In the background, every
`setup.sh` runs at once through `tasks.Start`. Tasks marked `order: last` then run one by one, because an
etcd snapshot or a broken scheduler or kubelet would spoil the setups beside it. Each task turns `ready` or
`failed`, and the file is saved after each. When all are done, the lock is released and the clock starts
(`started`, and `deadline` = started + minutes), so setup time is not exam time.

**Taking it.** The page loads through `showExam` (`apps/web/src/exam.ts`), which asks `GET /api/exam`
first. A 404 means no exam, so it shows the practice list. While the exam is preparing, the page shows
"N of M tasks set up" and asks again every 2 seconds. Then you see:
- A countdown. It is timed against the `now` the server sends, so a wrong browser clock doesn't matter. It
  turns red under 10 minutes, and after 0:00 it shows "Time's up +…", but nothing ends on its own.
- Question pills. A flagged question shows ⚑, and a failed setup shows in red with its error on the question.
- **Flag for later**, which sends `PUT` or `DELETE /api/exam/flags/{id}`.
- **End exam**.

You work in the same terminal (`ssh <host>`, then `k …`). While the exam runs, practice start, check and
solution answer 409 "not during an exam". Questions stay readable.

**Restart.** `cmd/server/main.go` opens the session from `exam.json`. If the file names a task the catalog
no longer has, the server refuses to start and says which file to delete. Any task still `preparing` is set
up again, with the lock held (`resumeExam`). A reload finds the same exam, with the same deadline and flags.

**Ending.** **End exam** asks for confirmation, then sends `POST /api/exam/end`. With the lock held, every
`check.sh` runs at once through `tasks.Check`. A check that crashes counts as 0 and keeps its error. Each
`grader.Result` and the end time are saved. The reply includes the score, worked out on the server by
`exam.Score`: each task's earned/total times its weight, over the sum of the weights, passing at 66%. The
results page shows the percentage with PASS or FAIL and one row per task with its points. Opening a row
shows the ✓/✗ checks and the solution, which is allowed now that the exam is over.

**Back to practice.** This sends `DELETE /api/exam`, which removes `exam.json`, and the practice list
comes back.
