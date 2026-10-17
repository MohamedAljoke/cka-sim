NS=tr-static-yaml
mirror=node-exporter-cka-sim-worker2
runs() { eq "file Running" "$(kubectl -n $NS get pod $mirror -o jsonpath='{.metadata.annotations.kubernetes\.io/config\.source} {.status.phase}' 2>/dev/null)"; }
kept() { runs && eq "busybox:1.36" "$(kubectl -n $NS get pod $mirror -o jsonpath='{.spec.containers[0].image}')" && kubectl -n $NS get pod $mirror -o jsonpath='{.spec.containers[0].command}' | grep -q loadavg; }
wait_for 60 runs

check 3 "Static Pod node-exporter runs on cka-sim-worker2" runs
check 1 "It keeps its image and command" kept
