NS=nw-external
selectorless() { kubectl -n $NS get service legacy-db >/dev/null 2>&1 && [ -z "$(kubectl -n $NS get service legacy-db -o jsonpath='{.spec.selector}')" ]; }
slice() { kubectl -n $NS get endpointslice legacy-db-1 -o jsonpath="$1"; }
slice_ok() { eq legacy-db "$(slice '{.metadata.labels.kubernetes\.io/service-name}')" && eq "$(cat "$COURSE/backend-ip")" "$(slice '{.endpoints[0].addresses[0]}')" && eq 8080 "$(slice '{.ports[0].port}')"; }
serves() { eq "legacy db api" "$(kubectl -n $NS exec client -- wget -qO- -T 3 http://legacy-db.$NS 2>/dev/null)"; }
working() { selectorless && serves; }
slice_ok && wait_for 30 serves

check 1 "Service legacy-db exists and has no selector" selectorless
check 2 "EndpointSlice legacy-db-1 belongs to legacy-db and points at the backend IP on 8080" slice_ok
check 3 "http://legacy-db.nw-external reaches the legacy API" working
