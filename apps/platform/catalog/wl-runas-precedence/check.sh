NS=wl-runas-precedence
P() { kubectl -n $NS get pod two -o jsonpath="$1" 2>/dev/null; }
spec_ok() { eq "1000 2000" "$(P '{.spec.securityContext.runAsUser} {.spec.containers[?(@.name=="sidecar")].securityContext.runAsUser}')" && [ -z "$(P '{.spec.containers[?(@.name=="app")].securityContext.runAsUser}')" ]; }
running() { eq True "$(P '{.status.conditions[?(@.type=="Ready")].status}')"; }
uid() { kubectl -n $NS exec two -c "$1" -- id -u 2>/dev/null; }
answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/uids.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
runs_as() { running && eq "1000 2000" "$(uid app) $(uid sidecar)"; }
answers() { runs_as && eq "1000 2000" "$(answer app) $(answer sidecar)"; }
P '{.metadata.name}' >/dev/null && wait_for 60 running

check 2 "Pod two sets runAsUser 1000 at Pod level and 2000 only on sidecar" spec_ok
check 1 "app runs as 1000 and sidecar as 2000" runs_as
check 2 "uids.txt gives both UIDs" answers
