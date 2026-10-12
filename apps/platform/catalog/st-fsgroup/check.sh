NS=st-fsgroup
P() { kubectl -n $NS get pod reader -o jsonpath="$1" 2>/dev/null; }
settings() { eq "1000 2000 288" "$(P '{.spec.securityContext.runAsUser} {.spec.securityContext.fsGroup} {.spec.volumes[0].secret.defaultMode}')"; }
reads() { settings && eq True "$(P '{.status.conditions[?(@.type=="Ready")].status}')" && eq hunter2 "$(kubectl -n $NS logs reader 2>/dev/null)"; }
not_root() { reads && eq 1000 "$(kubectl -n $NS exec reader -- id -u)"; }
settings && wait_for 60 reads

check 2 "reader runs as 1000 with fsGroup 2000 and defaultMode 0440" settings
check 2 "reader reads the password" reads
check 1 "It still runs as user 1000" not_root
