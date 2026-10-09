#!/usr/bin/env bash
# try-task.sh <task-id>: prove a catalog task works on the running cluster.
# Runs setup, check (must fail), solution, check (must fully pass), the same way the server runs
# scripts: lib.sh + the script, piped into bash on the task's host node.
set -euo pipefail

id=${1:?usage: try-task.sh <task-id>}
catalog="$(cd "$(dirname "$0")/../catalog" && pwd)"
dir="$catalog/$id"
[ -f "$dir/task.md" ] || { echo "no task at $dir" >&2; exit 1; }

host=$(sed -n 's/^host: *//p' "$dir/task.md" | head -1)
docker inspect "$host" >/dev/null 2>&1 || { echo "node $host is not running: make up" >&2; exit 1; }

run() {
  echo "── $1"
  cat "$catalog/lib.sh" "$dir/$1" | docker exec -i -e TASK_ID="$id" "$host" bash -s
}

# grade <must>: print the check output, then confirm it has no PASS (before) or no FAIL (after).
# Same rule as tasks.Selftest; replace this script once a CLI command calls that.
grade() {
  local out
  out=$(run check.sh)
  echo "$out"
  grep -q '^\(PASS\|FAIL\) ' <<<"$out" || { echo "✗ check.sh prints no checks" >&2; exit 1; }
  if [ "$1" = fails ] && grep -q '^PASS ' <<<"$out"; then
    echo "✗ check.sh gives points before the fix: every check must fail right after setup" >&2
    exit 1
  fi
  if [ "$1" = passes ] && grep -q '^FAIL ' <<<"$out"; then
    echo "✗ check.sh still fails after solution.sh" >&2
    exit 1
  fi
}

run setup.sh
grade fails
run solution.sh
grade passes
echo "✓ $id works: setup → check fails → solution → check passes"
