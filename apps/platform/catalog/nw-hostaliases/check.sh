NS=nw-hostaliases
running() { eq True "$(kubectl -n $NS get pod legacy -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null)"; }
spec_ok() { eq '10.10.0.5 cache.legacy.local db.legacy.local' "$(kubectl -n $NS get pod legacy -o json | jq -r '.spec.hostAliases[]? | select(.ip=="10.10.0.5") | [.ip] + (.hostnames | sort) | join(" ")')"; }
hosts_ok() { running && kubectl -n $NS exec legacy -- grep -E '^10\.10\.0\.5[[:space:]]' /etc/hosts | grep -q db.legacy.local && kubectl -n $NS exec legacy -- grep -E '^10\.10\.0\.5[[:space:]]' /etc/hosts | grep -q cache.legacy.local; }
kubectl -n $NS get pod legacy >/dev/null 2>&1 && wait_for 60 running

check 2 "legacy has hostAliases for both names at 10.10.0.5" spec_ok
check 2 "Both names are in legacy's /etc/hosts" hosts_ok
