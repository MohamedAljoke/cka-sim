# Helpers sourced before every setup.sh, check.sh and solution.sh.
# These scripts run on your machine, not on base: they drive the clusters with kubectl and
# reach into nodes with docker exec. Environment: CLUSTER, CONTEXT, HOST, TASK_ID, TASK_DIR.

# kubectl against the task's cluster.
k() { kubectl --context "$CONTEXT" "$@"; }

# Run a command as root on a node (a container named after the node).
on() {
  local node=$1
  shift
  docker exec -i "$node" bash -c "$*"
}

# Run a command as root on the task's host.
on_host() { on "$HOST" "$@"; }

# check <points> <description> <command...>
# Prints PASS or FAIL for the grader. The command's own output is discarded.
check() {
  local points=$1 description=$2
  shift 2
  if "$@" >/dev/null 2>&1; then
    echo "PASS $points $description"
  else
    echo "FAIL $points $description"
  fi
}

# eq <expected> <actual> — a readable equality test for check.
eq() { [ "$1" = "$2" ]; }

# Retry a command until it succeeds or the timeout (seconds) passes.
wait_for() {
  local timeout=$1
  shift
  local deadline=$((SECONDS + timeout))
  until "$@" >/dev/null 2>&1; do
    [ $SECONDS -ge $deadline ] && return 1
    sleep 2
  done
}

# Fresh namespace for a task: delete leftovers from an earlier attempt, then create it.
# Pods are force-deleted first: a Pod on a node whose kubelet another task broke can never
# confirm its own termination, and would hold the namespace in Terminating forever.
fresh_ns() {
  k -n "$1" delete pods --all --force --grace-period=0 >/dev/null 2>&1
  k delete namespace "$1" --ignore-not-found --wait=true --timeout=120s >/dev/null
  k create namespace "$1" >/dev/null
}

# A file in the task's exam directory on the host, /opt/course/<task id>/...
course_dir() { echo "/opt/course/$TASK_ID"; }

# remember <key> <value> / recall <key>: state a setup hands to its check (UIDs, generations),
# kept on your machine under ~/.cache/cka-sim/state so it is invisible from the exam hosts.
_state_file() { echo "${CKA_SIM_STATE:?}/$TASK_ID.$1"; }
remember() { mkdir -p "$CKA_SIM_STATE" && printf '%s' "$2" > "$(_state_file "$1")"; }
recall() { cat "$(_state_file "$1")" 2>/dev/null; }
