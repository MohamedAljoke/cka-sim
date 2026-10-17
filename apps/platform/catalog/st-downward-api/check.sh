NS=st-downward-api
running() { eq True "$(kubectl -n $NS get pod meta -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null)"; }
in_pod() { kubectl -n $NS exec meta -- cat "/etc/podinfo/$1" 2>/dev/null; }
labels_ok() { running && in_pod labels | grep -qx 'app="meta"'; }
annotations_ok() { running && in_pod annotations | grep -qx 'owner="team-a"'; }
mem_ok() { running && eq 64 "$(in_pod mem_limit)"; }
kubectl -n $NS get pod meta >/dev/null 2>&1 && wait_for 60 running

check 1 "labels holds app=meta" labels_ok
check 1 "annotations holds owner=team-a" annotations_ok
check 2 "mem_limit holds 64" mem_ok
