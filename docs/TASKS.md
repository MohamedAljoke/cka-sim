# Tasks: set up, graded and healed on a real cluster

Every one of the 203 tasks is a folder of bash scripts the lab runs on a real node. A task breaks or prepares the
cluster, grades what you did by looking at the cluster itself, and puts back whatever it broke, so the next
task starts on a healthy cluster without a rebuild.

| Domain | Tasks | Exam weight |
| --- | --- | --- |
| Troubleshooting | 36 | 30% |
| Cluster Architecture, Installation and Configuration | 46 | 25% |
| Services and Networking | 52 | 20% |
| Workloads and Scheduling | 47 | 15% |
| Storage | 22 | 10% |

**Contents:** [1. A task folder](#1-a-task-folder) · [2. How a script runs](#2-how-a-script-runs) ·
[3. Grading](#3-grading) · [4. One cluster, many tasks](#4-one-cluster-many-tasks) ·
[5. Selftest](#5-selftest-every-task-solved-for-real) · [6. The exam](#6-the-exam)

## 1. A task folder

`apps/platform/catalog/<id>/`, embedded in the platform binary:

| File | What it is |
| --- | --- |
| `task.md` | Front matter (`id`, `title`, `domain`, `host`, `topics`, `weight`, optional `order: last`) and the question |
| `explain.md` | The solution, shown on the page when you ask for it |
| `setup.sh` | Breaks or prepares the cluster |
| `check.sh` | Grades it |
| `solution.sh` | Does what `explain.md` says, for the selftest |
| `reset.sh` | Only for the 25 tasks that break the cluster itself (a kubelet, etcd, the scheduler, CoreDNS…): undoes the setup, and does nothing on a healthy cluster |

The smallest task, `wl-pod`:

```sh
# setup.sh
fresh_ns wl-pod

# check.sh
pod() { kubectl -n wl-pod get pod web -o jsonpath="$1" 2>/dev/null; }
running() { eq Running "$(pod '{.status.phase}')"; }

check 1 "Pod web uses image nginx:1.27" eq nginx:1.27 "$(pod '{.spec.containers[0].image}')"
if pod '{.metadata.name}' >/dev/null; then
  check 1 "Pod web is Running" wait_for 60 running
else
  echo "FAIL 1 Pod web is Running"
fi
```

A troubleshooting task goes much further. `tr-kubelet`'s setup points the kubelet's systemd drop-in at a binary
that doesn't exist, disables the service, and waits until the worker is NotReady. Its reset puts the drop-in back
and waits until the worker is Ready again.

## 2. How a script runs

The platform never runs a script itself. It sends the script to the agent with `RunScript`, naming the task's
host:

- **What is sent.** The loader (`internal/infra/adapters/files/task_files.go`) prefixes every script with
  `export TASK_ID=<id>` and the catalog's `lib.sh`, so the lab needs nothing of its own. `lib.sh` holds the shared
  helpers: `check`, `eq`, `wait_for`, `fresh_ns`, `connects`, `fetch`, `fresh_course` for `/opt/course/<id>/` answer
  files, and `etcd_tools`.
- **Where it runs.** The agent accepts only a host from the cluster's nodes, never the base host. It runs
  `docker exec -i <node> setsid -w bash -s` with the script on stdin, so it runs as root, with an admin kubeconfig.
- **When it's cut off.** `setsid` makes the script lead its own process group. A deadline kills that whole
  group inside the node, not only the `docker` client, which would leave the script running.
- **What comes back.** The exit code and the last 64 KiB of output, the end being where a failure says why.

Each use case sets a limit: 3 minutes for a setup, a reset or a solution, 2 for a check, 1 for a tidy. Setups and
resets run to the end even if the page that asked goes away, because half a setup leaves the cluster broken.

## 3. Grading

A check prints one line per thing it checks, and anything else is ignored:

```
PASS 1 Pod web uses image nginx:1.27
FAIL 1 Pod web is Running
```

`lib.sh`'s `check <points> <description> <command…>` runs the command and prints the line, throwing the
command's own output away. `tasks.ReadGrade` (`internal/domain/tasks/grade.go`) adds the lines up into earned and
total. A check that prints no line at all broke, and the page shows its output.

## 4. One cluster, many tasks

Practice and exams share one cluster, so a task must not leave a mess for the next one.

- **Namespaces.** `fresh_ns` recreates a task's namespace and labels it `cka-sim/task=<id>`. When you leave a
  task you set up, the page asks for a **tidy**, which deletes every namespace with that label.
- **Touched tasks.** A setup marks its task touched. Before every practice start, and before an exam's setups,
  **heal** runs the reset of every touched task, one at a time. A reset that fails leaves its task touched, so
  the next heal tries it again.
- **The script lock.** Setups take a read lock, so several can run at once, as an exam needs. A reset or a tidy
  takes the write lock, so it runs alone, and a setup can't land between two of them.

The touched list lives in memory, so a restarted platform forgets what practice broke. That's why a tidy runs
the reset even for a task that isn't marked touched.

## 5. Selftest: every task solved for real

The checks are tests too. If doing nothing earns points, or the solution can't reach full marks, the task is
broken. `make selftest` (or `make selftest TASK=<id>`) runs each task against a Ready lab:

1. start: heal, then `setup.sh`;
2. check, which must earn **0** with a total above 0;
3. `solution.sh`;
4. check, which must earn **full** marks;
5. tidy.

The use case is `tasks/usecases.SelftestTask`. The test, `catalog/selftest_test.go`, only runs with
`-selftest`. Today 198 of 203 tasks pass. `tr-events` earns 3 of 4 after its solution, and the four `wl-hpa-*`
checks take longer than the 2-minute check limit.

## 6. The exam

Like killer.sh: a fixed number of tasks against the clock, graded at the end.

- **Draw.** `tasks.Draw` picks tasks so the domain mix follows the weights above, by largest remainders, and a
  domain without enough tasks hands its share to the heaviest ones that have some left.
- **Order.** At most one task is drawn from the 20 marked `order: last`, such as an etcd restore or a broken
  scheduler, because those would undo or block the other setups. It is set up after the others and shown
  first.
- **Setup, then the clock.** The exam is healed first. The setups run in the background, and the clock starts
  only once every setup is done.
- **Saved and resumed.** The exam is saved to `<state>/exam.json`. A restarted platform resumes its setups and
  checks, and an exam read past its deadline ends.
- **Score.** Each task's share of its own points is weighted by the task's weight, and 66% passes. A task whose
  setup failed counts in neither earned nor total. A scored exam tidies all its tasks, and the next exam waits
  for that tidy.
