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

A new cluster then gets the task images (`Cluster.LoadImages`, `internal/cluster/images.go`): `up` pulls
any of `cluster.Images` this machine's Docker lacks, then pipes `docker save` into each node's `ctr images
import`. That adds about 8 s, and the first exam doesn't wait on Docker Hub on every node. If it fails, `up`
only warns, and the first exam pulls the images itself. `catalog_test.go` fails when a script uses an image
the list lacks.

Next, `make dev` builds `bin/server` and starts it in the background, then starts Vite on 5173.
`vite.config.ts` forwards `/api` and `/ws` to the Go server on `127.0.0.1:7070`. The server
(`cmd/server/main.go`) claims the port, loads the catalog (`tasks.Load(catalog.FS)`: every
`catalog/<id>/task.md` and its scripts, built into the binary) and `lib.sh`, and makes a **lab** with the
`sandbox.Local` provider (`internal/sandbox`). It doesn't touch the cluster yet; that happens when a lab
starts (flow 2). Then it serves.

Ctrl-C reaches the server directly, because it runs as a binary and not through `go run`. The Makefile's
trap waits for it, and on the way out the server ends the lab, which closes every page shell.

## 2. Opening the page: home, the lab and the terminal

**Home.** `main.ts` first asks `GET /api/exam` (`showExam`, `src/exam.ts`). A running or preparing exam
opens straight away (flow 6). With no exam, or a finished one, it shows **Home** (`src/home.ts`), full width
with no terminal. Home needs no lab, because the catalog and the exam settings come from the server alone.
`GET /api/tasks` returns every task's metadata (title, domain and its `domainTitle`, topics, weight, host),
never the scripts. Home has, top to bottom:
- **Filters.** Two multi-selects, **Domains** and **Topics** (`src/multiselect.ts`), with everything
  chosen at first; a long list gets a search box. They narrow both the exam and the practice list, and
  "3 of 8 tasks" says how much is left.
- **Mock exam.** The number of tasks and the minutes; the minutes default to 7.5 per task, so 16 tasks get
  the CKA's 120. With nothing narrowed it draws from the whole catalog in the CKA domain mix; otherwise
  from the filtered tasks, so choosing only the `rbac` topic gives an RBAC exam.
- **Practise a task.** The filtered catalog.
- **Last exam.** "Last exam: 62% · FAIL" with **View results**, above the rest, when there is a finished exam.

**The lab.** A lab is the box your cluster lives in. Locally that's the kind cluster `make up` made; on
Fly each user gets a VM instead (flow 7), behind the same `sandbox.Provider` interface. There's no
Start button. Practising a task or starting an exam starts it on the server. `Lab.Start`
(`internal/sandbox/lab.go`) calls the provider's `Create` in the background. `Local.Create`
(`internal/sandbox/local.go`) checks that the cluster and base are up (`cluster.RequireUp`), closes any
shells left over from a crashed server (`EndAll`), and returns a box: the runner and the `DockerOpener` for
this cluster.

The topbar (`src/lab.ts`) shows nothing while there is no lab. When a view tells it the server may have
started one (`app.labChanged`), it asks `GET /api/lab` every second while the state is `starting` ("Starting
lab on Local Docker… 0.2s"). Then it shows "Lab ready in 0.3s · Local Docker" with **End lab**, or the error.
The `Lab` is itself the runner and the opener the rest of the server uses, so every script and shell goes
through it to the current box. **End lab** sends `DELETE /api/lab` (409 while an exam runs), which takes the
script lock and calls `Destroy`. Locally that closes the shells and keeps the cluster.

**The terminal.** Practice and exam views are split: the question on the left and the terminal on the right
(`app.split`, `src/main.ts`). Home and results are full width (`.exam.single` hides the terminal pane). The
first time a split view finds the lab ready, `main.ts` opens a websocket to `/ws/terminal`. The server's
`DockerOpener` (`internal/terminal/docker.go`) answers with the equivalent of
`docker exec -it -u candidate -w /home/candidate cka-sim-base bash` and pipes bytes both ways between xterm
and that shell. So you land on `candidate@base` with no kubectl, as in the exam. The terminal stays open
while you move between views, and closes when the lab ends. When the tab closes, the shell is hung up as the
same user; root can't do it, because reading another user's process environment needs ptrace, which
Docker withholds.

## 3. Clicking a task: question and setup

Clicking a task on Home opens **practice** (`showPractice`, `src/practice.ts`), split screen, and sends two
requests in parallel:

- `GET /api/tasks/{id}/question` returns the question markdown, and `marked` renders it right away.
- `POST /api/tasks/{id}/start` reaches `startTask` (`internal/server/api.go`). It first waits for the lab
  (`ensureLab` → `Lab.Ensure`), starting one if there is none; the page meanwhile tells the topbar to show
  it, and the terminal opens once it's ready. A lab that fails answers 503. It then takes the setup lock; if
  a setup is already running, it answers 409. It first calls `tasks.Heal` (`internal/tasks/run.go`), which runs
  the `reset.sh` of every task that ships one. These put back what a task left unsolved broke outside its
  namespace, such as `tr-kubelet`'s kubelet, and do nothing on a healthy cluster. It then calls `tasks.Start`,
  which reads `setup.sh` and hands it to the runner with a 3-minute limit.

The lab passes the script to its box's runner. The local one (`internal/runner/runner.go`) runs `docker exec -i -e TASK_ID=<id> <task's host> bash -s` and
writes `lib.sh` followed by `setup.sh` to its stdin. The script runs as root on the node the task names,
using root's kubeconfig. Helpers such as `fresh_ns` wipe and recreate the task's namespace before the
script builds the broken state. **Home** or **Reset** aborts the request, and the runner then kills the
script's whole process group inside the node (it runs under `setsid`, so its `kubectl` and `sleep` go too).
That frees the lock, and the next start waits up to 60 seconds for it. A half-built task is harmless,
because every setup starts from `fresh_ns` and the resets. `fresh_ns` also labels the namespace
`cka-sim/task=<id>`, so the task can be tidied later.

The page shows "Preparing the cluster (the lab starts first if there is none)…", then "Ready" on 204, or the script's output if setup failed. You
then follow the "Connect first" line: `ssh <host>` logs you in with base's key and no password, and you
solve the task with `k`.

**Leaving.** **Home**, another task, or closing the tab sends `POST /api/tasks/{id}/tidy` (`keepalive`, so
it survives the tab closing). `tidyTask` answers 204 when there is no lab, 409 during an exam, and
otherwise 202 at once. In the background, under the script lock, `tasks.Tidy` runs the task's `reset.sh`
and `tasks.TidyScript`, which deletes the namespaces labelled for that task without waiting for them to go.
The next setup's `fresh_ns` then finds nothing to delete. If the tidy never ran (a crash), the heal and
`fresh_ns` before the next setup still clean up, only slower.

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
(`tasks.Solve`) and checks again, which must earn every point. The task's `reset.sh` runs before setup and
again when the subtest ends, so one failing task can't break the node for the next. A failure names the task
and the checks at fault, for example:
`wl-scale: earns 0/4 after the solution: "Deployment web wants 4 replicas", …`.

## 6. Exam mode

**Starting.** On Home, the **Mock exam** card (`apps/web/src/home.ts`) sends
`POST /api/exam {count, minutes, domains, topics}`. A multi-select with everything chosen sends an empty
list, which means any; otherwise it sends what you chose. `beginExam` (`internal/server/exam.go`) checks the numbers (400 if they are out of range) and
narrows the catalog with `tasks.Filter` (`internal/tasks/draw.go`): a task must be in one of the domains
and have one of the topics, where an empty list means any. No match gives 400. It takes the script lock,
so a running practice setup gets 409, and starts the lab if there is none. Then it calls
`exam.Session.Begin` (`internal/exam/session.go`) with that pool. A finished, scored exam is replaced; one
still being scored gives 409 "the last exam is still being scored". `tasks.Draw` picks the tasks:
each domain gets its share of the count by the CKA weights, with any shortfall moved to the heaviest domains,
and the result is shuffled. The exam is saved with every task `preparing` to
`~/.local/state/cka-sim/exam.json`, and the request answers 202 right away. In the background, setup first waits for the lab (`Config.Ready`,
which is `Lab.Ensure`); if the lab fails, every task is marked `failed` with its error. Then `tasks.Heal`
first runs every `reset.sh` (on a restart, not those of tasks this exam already set up). Then every
`setup.sh` runs at once through `tasks.Start`. When those are done, the clock starts (`started`, and
`deadline` = started + minutes). Tasks marked `order: last` then run in the background, because an etcd
snapshot or a broken scheduler or kubelet would spoil the setups beside it: one by one per host, and hosts
side by side, so `tr-kubelet` on a worker doesn't wait for `ar-etcd` and `tr-scheduler` on the control
plane. You read and work meanwhile. When the last one is done, the lock is released and `deadline` moves out by the time they took,
so setup time is never exam time. Each task turns `ready` or `failed`, and the file is saved after each.
The server log shows how long each setup, the heal, each check and the tidy took.

**Taking it.** The page loads through `showExam` (`apps/web/src/exam.ts`), which asks `GET /api/exam`
first. A 404 means no exam, so it shows Home. The exam is split screen, with the terminal on the right. While the exam is preparing, the page lists
each task as "setting up…", "ready" or "failed" and asks `GET /api/exam` again every 2 seconds (the catalog
is loaded once). Then you see:
- A countdown. It is timed against the `now` the server sends, so a wrong browser clock doesn't matter. It
  turns red under 10 minutes, and after 0:00 it shows "Time's up +…", but nothing ends on its own.
- Question pills numbered 01, 02, 03. A flagged question shows ⚑, and a failed setup shows in red with its error on the question.
  A task still setting up has a dashed "05 …" pill and shows "still being set up" instead of its question.
  While any is, the page keeps polling and picks up each task as it turns ready, and the moved deadline.
- **Flag for later**, which sends `PUT` or `DELETE /api/exam/flags/{id}`.
- **End exam**, disabled until every task is set up (`Session.End` refuses too).

You work in the same terminal (`ssh <host>`, then `k …`). While the exam runs, practice start, check and
solution answer 409 "not during an exam". Questions stay readable.

**Restart.** `cmd/server/main.go` opens the session from `exam.json`. If the file names a task the catalog
no longer has, the server refuses to start and says which file to delete. Any task still `preparing` is set
up again, with the lock held (`resumeExam`), and that setup starts the lab itself; a clock that already
started keeps its start. An exam that ended but
wasn't fully scored has its missing checks run again the same way. A reload finds the same exam, with the
same deadline and flags.

**Ending.** **End exam** asks for confirmation, then sends `POST /api/exam/end`. `endExam` takes the script
lock and calls `Session.End`, which saves `ended` and answers 202 at once; you are never stuck on
"Scoring…". In the background, `grade` runs every `check.sh` at once through `tasks.Check`, saving each
task's `grader.Result` as it lands. A check that crashes counts as 0 and keeps its error. Checks that wait for
the cluster to settle only wait when the wait can succeed: `tr-kubelet` when the kubelet is running,
`ar-etcd` when etcd or the API server is missing or started in the last 2 minutes, `tr-scheduler` when a
scheduler container runs, `st-pvc` when Pod `writer` exists, `wl-rollout` when the Deployment is back on the
good image, `tr-service` when the Service has endpoints. So an untouched task scores in under 3 seconds
instead of up to 150. Then `tasks.Tidy`
runs the exam's `reset.sh` scripts, so a kubelet left broken doesn't outlive the exam, and deletes the
exam's task namespaces, so the next exam's setups find nothing to delete. Last, it saves
`scored`, ends the lab (`freeLab`) and frees the lock: locally that only closes the shells and the terminal,
on Fly it destroys the exam's VM (flow 7). The score is worked out on the server by `exam.Score`, sent as soon as every
task has its result (`Exam.Graded`), without waiting for the tidy: each task's earned/total times its weight,
over the sum of the weights, passing at 66%.

**Results.** The page switches to full width and asks `GET /api/exam` every 2 seconds until `scored`. Each
row reads "01 Scale a Deployment" with "checking…" until its check is back, then `2 / 4`. A row with
points missing starts open and red: its ✓/✗ checks, the solution, and **Try again**, which opens that task
in practice once scoring is done. The percentage with PASS or FAIL appears when every row is in; until the
tidy finishes, "Tidying up the cluster…" shows under it and **Try again** stays disabled. **Home**
keeps the results; Home shows them as "Last exam" until the next exam replaces them.

## 7. On Fly: Start exam gets you a VM of your own

**Setting it up (once, from `apps/cka-sim/deploy/fly`).** `make apps` creates two Fly apps: the site
(`cka-sim-maljoke`) and the pool (`cka-sim-labs-maljoke`, no public address). It stores the site's secrets:
a Fly token for the pool app and a password. `make image` builds `lab.Dockerfile` (dind plus the `cka-sim`
CLI) and pushes it. `make pool N=2` runs `pool.sh fill`: for each VM it does `fly machine run` with
2 performance CPUs and 4 GB, labels it `pool=building`, and runs `lab-build` in it over `fly ssh`. That waits
for dockerd and runs `cka-sim up`. Then `pool.sh` sets `pool=ready` and suspends the VM (~2.5 min each, in
parallel). `lab-start.sh`, the VM's entrypoint, works around three Fly quirks before dockerd starts:
- an ext4 loop disk at `/var/lib/docker`, because the root disk is an overlay and Docker would fall back to `vfs`;
- a `name=systemd` cgroup, without which systemd in the kind nodes exits;
- `--dns 8.8.8.8`, because Fly's resolver is IPv6 only and image builds can't reach it.

It also has dockerd listen on `tcp://[::]:2375`, reachable only on Fly's private network. `make deploy`
builds `server.Dockerfile` (the built page, the server, and the `docker` CLI) and deploys it with `fly.toml`.

**Opening the site.** `https://cka-sim-maljoke.fly.dev` asks for the password: `server.RequirePassword`
wraps everything, and the server refuses to listen beyond loopback without `CKA_SIM_PASSWORD`.
`server.Site` serves the built page at `/` beside `/api` and `/ws`, so it is one origin. `cmd/server/main.go`
reads `CKA_SIM_PROVIDER=fly` and builds `fly.New` (`internal/fly`), the only Fly-aware code. It first
destroys any VM labelled `pool=claimed`, which a previous run took and never gave back. The topbar shows
nothing until a lab starts.

**Starting an exam (or practising a task).** The lab starts as in flow 2, but `fly.Provider.Create` runs these steps:
1. List the pool app's Machines with `metadata.pool=ready`.
2. Label the first one `claimed`.
3. Resume it (`POST /machines/{id}/start`) and wait for `started`.
4. Hand its private address to `sandbox.DockerBox` as `tcp://[fdaa:…]:2375`. The same `runner.Node`
   (`docker exec`) and `DockerOpener` as locally now talk to that VM's dockerd.
5. Health check: `kubectl get --raw=/readyz` on the control plane, retried for 30s. If Fly cold-booted
   the VM instead of restoring its snapshot, the cluster is gone, so the VM is destroyed and the next one is tried.

The topbar shows "Lab ready in 4.0s · Fly". From here, setup scripts, checks and the terminal all run on
that VM.

**Ending.** End exam grades and tidies, then ends the lab, and `Destroy` deletes the Machine (`?force=true`). End lab
and `DELETE /api/exam` do the same. The pool is now one smaller. Nothing refills it yet: run `make pool N=1`. With an empty pool,
the lab fails with "no warm VM in the pool", shown in the topbar.
