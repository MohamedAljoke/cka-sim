scheduled() { [ -n "$(k -n sched-check get pod probe -o jsonpath='{.spec.nodeName}')" ]; }
running() { eq Running "$(k -n sched-check get pod probe -o jsonpath='{.status.phase}')"; }
scheduler_ready() { eq True "$(k -n kube-system get pod kube-scheduler-cka3962-control-plane -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
same_probe() { running && eq "$(recall probe-uid)" "$(k -n sched-check get pod probe -o jsonpath='{.metadata.uid}')"; }
wait_for 90 running
wait_for 60 scheduler_ready
check 2 "kube-scheduler static Pod is Ready" scheduler_ready
check 3 "Pod probe was scheduled to a node" scheduled
check 2 "Pod probe is Running" running
check 1 "Pod probe was not recreated" same_probe
