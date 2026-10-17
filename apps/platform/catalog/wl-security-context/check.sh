NS=wl-security-context
running() { eq True "$(kubectl -n $NS get pod secure -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
in_pod() { kubectl -n $NS exec secure -- "$@"; }
ids() { running && eq "1000 3000" "$(in_pod id -u) $(in_pod id -g)"; }
readonly_root() { running && ! in_pod touch /probe && in_pod touch /tmp/probe; }
sc() { kubectl -n $NS get pod secure -o jsonpath="{.spec.containers[0].securityContext.$1}"; }
locked() { running && eq '["ALL"]' "$(sc capabilities.drop)" && eq false "$(sc allowPrivilegeEscalation)" && in_pod grep -q '^CapEff:.*0000000000000000' /proc/1/status; }
kubectl -n $NS get pod secure >/dev/null 2>&1 && wait_for 60 running

check 2 "secure runs as UID 1000 and GID 3000" ids
check 2 "The root filesystem is read-only and /tmp is writable" readonly_root
check 2 "All capabilities are dropped and privilege escalation is off" locked
