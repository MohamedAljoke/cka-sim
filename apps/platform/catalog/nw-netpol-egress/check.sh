NS=nw-netpol-egress
DB=$(pod_ip $NS db)
CACHE=$(pod_ip $NS cache)
db_port() { connects $NS backend "$DB" 5432 && ! connects $NS backend "$DB" 3306; }
no_cache() { ! connects $NS backend "$CACHE" 6379; }
others_free() { connects $NS frontend "$CACHE" 6379 && connects $NS frontend "$DB" 3306 && ! connects $NS backend "$CACHE" 6379; }
selects_backend() { eq "backend Egress" "$(kubectl -n $NS get networkpolicy backend-egress -o jsonpath='{.spec.podSelector.matchLabels.app} {.spec.policyTypes[*]}')"; }

check 2 "backend reaches db:5432 but not db:3306" db_port
check 1 "backend can't reach cache:6379" no_cache
check 1 "Only backend is restricted: frontend still reaches cache and db:3306" others_free
check 1 "backend-egress selects app=backend for Egress only" selects_backend
