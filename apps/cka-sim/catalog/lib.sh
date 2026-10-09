# Sourced before every setup.sh, check.sh and solution.sh. They run in the control-plane node,
# where kubectl is already admin. TASK_ID is set.

# check <points> <description> <command...>
# Prints the PASS/FAIL line the grader reads; the command's own output is thrown away.
check() {
  local points=$1 description=$2
  shift 2
  if "$@" >/dev/null 2>&1; then
    echo "PASS $points $description"
  else
    echo "FAIL $points $description"
  fi
}

eq() { [ "$1" = "$2" ]; }

# wait_for <seconds> <command...>: retry until it succeeds or time runs out.
wait_for() {
  local timeout=$1
  shift
  local deadline=$((SECONDS + timeout))
  until "$@" >/dev/null 2>&1; do
    [ "$SECONDS" -ge "$deadline" ] && return 1
    sleep 2
  done
}

# A pod on a node whose kubelet another task broke never confirms termination and would hold
# the namespace in Terminating forever, so pods are force-deleted first. Their controllers go
# before them, or they would recreate pods while the namespace empties (halves the wait).
fresh_ns() {
  kubectl -n "$1" delete deployments,replicasets,statefulsets,daemonsets,jobs,cronjobs --all --wait=false >/dev/null 2>&1
  kubectl -n "$1" delete pods --all --force --grace-period=0 >/dev/null 2>&1
  kubectl delete namespace "$1" --ignore-not-found --wait=true --timeout=120s >/dev/null
  kubectl create namespace "$1" >/dev/null
}
