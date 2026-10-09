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

## 4. Not built yet

- **Check:** runs `check.sh` on the same host and turns its `PASS|FAIL <points> <description>` lines into a
  score.
- **Solution:** shows `explain.md`.
- **Reset:** runs `start` again.
- **`make selftest`:** for every task, runs setup → check (must score 0) → solution → check (must score
  full).
- **Exam mode:** the same pieces, with tasks drawn by domain weight, set up together, and a timer.
