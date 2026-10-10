pod() { kubectl -n sched-check get pod probe -o jsonpath="$1"; }
scheduled() { [ -n "$(pod '{.spec.nodeName}')" ]; }
running() { eq Running "$(pod '{.status.phase}')"; }
scheduler_ready() { eq True "$(kubectl -n kube-system get pod kube-scheduler-cka-sim-control-plane -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
# A recreated probe would be younger than the scheduler that finally placed it.
created_before_fix() {
  local started
  started=$(crictl inspect "$(crictl ps -q --name '^kube-scheduler$' | head -1)" | grep -m1 '"startedAt"' | cut -d'"' -f4)
  [ "$(date -d "$(pod '{.metadata.creationTimestamp}')" +%s)" -lt "$(date -d "$started" +%s)" ]
}
# A scheduler that isn't running yet (still the broken command) will never place the probe.
if crictl ps -q --name '^kube-scheduler$' | grep -q .; then
  wait_for 60 scheduler_ready
  wait_for 90 running
fi

check 2 "kube-scheduler static Pod is Ready" scheduler_ready
check 3 "Pod probe was scheduled to a node" scheduled
check 2 "Pod probe is Running" running
not_recreated() { running && created_before_fix; }
check 1 "Pod probe was not recreated" not_recreated
