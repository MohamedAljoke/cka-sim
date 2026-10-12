schedulable() { [ "$(kubectl get node cka-sim-worker2 -o jsonpath='{.spec.unschedulable}')" != true ]; }
on_worker2() { [ -n "$(kubectl -n tr-cordon get pods -l app=agent --field-selector spec.nodeName=cka-sim-worker2 -o name)" ]; }
generation() { kubectl -n tr-cordon get deployment agent -o jsonpath='{.metadata.generation}'; }
answer() { tr -d '[:space:]' < "$COURSE/node" 2>/dev/null; }

schedulable && wait_for 30 on_worker2
check 2 "Node cka-sim-worker2 is schedulable" schedulable
check 2 "A Pod of agent is scheduled on cka-sim-worker2" on_worker2
check 1 "/opt/course/tr-cordon/node names cka-sim-worker2" eq cka-sim-worker2 "$(answer)"
if on_worker2; then
  check 1 "Deployment agent was not modified" eq 1 "$(generation)"
else
  echo "FAIL 1 Deployment agent was not modified"
fi
