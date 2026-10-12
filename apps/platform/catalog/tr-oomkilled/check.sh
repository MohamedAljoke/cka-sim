NS=tr-oomkilled
c() { kubectl -n $NS get deployment cache -o jsonpath="{.spec.template.spec.containers[0].$1}"; }
reason_ok() { grep -qx 'OOMKilled' <(tr -d ' \r' < "$COURSE/reason.txt" 2>/dev/null); }
limit_ok() { eq 128Mi "$(c resources.limits.memory)"; }
D() { kubectl -n $NS get deployment cache -o jsonpath="$1"; }
stable() { limit_ok && eq 1 "$(D '{.status.updatedReplicas}')" && eq 1 "$(D '{.status.readyReplicas}')" && kubectl -n $NS logs deploy/cache | grep -q ready; }
kept() { stable && eq 16Mi "$(c resources.requests.memory)" && c command | grep -q 'head -c 50m'; }
limit_ok && wait_for 90 stable

check 1 "reason.txt says OOMKilled" reason_ok
check 2 "cache has a memory limit of 128Mi" limit_ok
check 2 "cache runs and finished loading" stable
check 1 "The command and the memory request are unchanged" kept
