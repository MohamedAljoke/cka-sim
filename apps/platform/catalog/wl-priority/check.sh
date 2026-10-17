NS=wl-priority
pc_ok() { eq 1000000 "$(kubectl get priorityclass critical -o jsonpath='{.value}')" && [ "$(kubectl get priorityclass critical -o jsonpath='{.globalDefault}')" != true ]; }
pod() { kubectl -n $NS get pod important -o jsonpath="$1" 2>/dev/null; }
runs() { eq "critical 1000000 Running cka-sim-worker2" "$(pod '{.spec.priorityClassName} {.spec.priority} {.status.phase} {.spec.nodeName}')"; }
request() { pod '{.spec.containers[0].resources.requests.cpu}'; }
kept() { runs && eq "$(request)" "$(kubectl -n $NS get deployment filler -o jsonpath='{.spec.template.spec.containers[0].resources.requests.cpu}')" && eq 1 "$(kubectl -n $NS get deployment filler -o jsonpath='{.metadata.generation}')"; }
pc_ok && wait_for 90 runs

check 2 "PriorityClass critical has value 1000000 and is not the global default" pc_ok
check 3 "Pod important runs on cka-sim-worker2 with priority critical" runs
check 1 "important kept its CPU request, and filler was not modified" kept
