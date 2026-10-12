NS=wl-resize
P() { kubectl -n $NS get pod worker -o jsonpath="$1" 2>/dev/null; }
desired() { eq "300m 400m" "$(P '{.spec.containers[0].resources.requests.cpu} {.spec.containers[0].resources.limits.cpu}')"; }
applied() { desired && eq "300m 400m" "$(P '{.status.containerStatuses[0].resources.requests.cpu} {.status.containerStatuses[0].resources.limits.cpu}')"; }
age() { echo $(( $(date -d "$(P '{.metadata.creationTimestamp}')" +%s) - $(date -d "$(kubectl get namespace $NS -o jsonpath='{.metadata.creationTimestamp}')" +%s) )); }
same() { applied && eq 0 "$(P '{.status.containerStatuses[0].restartCount}')" && [ "$(age)" -lt 60 ]; }
desired && wait_for 60 applied

check 2 "worker's spec asks for 300m CPU request and 400m limit" desired
check 2 "The running container has the new CPU resources" applied
check 2 "The Pod was neither recreated nor restarted" same
