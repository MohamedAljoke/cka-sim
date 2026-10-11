NS=wl-taints
W=cka-sim-worker
field() { kubectl -n $NS get pod "$1" -o jsonpath="$2"; }
tainted() { kubectl get node $W -o jsonpath='{range .spec.taints[*]}{.key}={.value}:{.effect}{"\n"}{end}' | grep -qx 'dedicated=gpu:NoSchedule'; }
tolerates() { [[ " $(field "$1" '{.spec.tolerations[*].key}') " == *" dedicated "* ]]; }
running() { eq Running "$(field "$1" '{.status.phase}')"; }
gpu_ok() { tolerates gpu-job && running gpu-job && eq $W "$(field gpu-job '{.spec.nodeName}')"; }
regular_ok() { ! tolerates regular && running regular && ! eq $W "$(field regular '{.spec.nodeName}')"; }

check 2 "cka-sim-worker has the taint dedicated=gpu:NoSchedule" tainted
# Without the taint any Pod can land on the worker, so the Pods only count once it's there.
if tainted; then
  check 3 "Pod gpu-job tolerates dedicated and runs on cka-sim-worker" wait_for 60 gpu_ok
  check 1 "Pod regular has no toleration and runs on another node" wait_for 60 regular_ok
else
  echo "FAIL 3 Pod gpu-job tolerates dedicated and runs on cka-sim-worker"
  echo "FAIL 1 Pod regular has no toleration and runs on another node"
fi
