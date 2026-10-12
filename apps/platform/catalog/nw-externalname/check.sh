NS=nw-externalname
svc() { kubectl -n $NS get service search -o jsonpath="$1" 2>/dev/null; }
alias_ok() { eq "ExternalName api.nw-externalname-search.svc.cluster.local" "$(svc '{.spec.type} {.spec.externalName}' | sed 's/\.$//')"; }
cname() { alias_ok && kubectl -n $NS exec client -- nslookup search.$NS.svc.cluster.local 2>/dev/null | grep -q 'canonical name = api.nw-externalname-search.svc.cluster.local'; }
works() { alias_ok && eq "search results" "$(kubectl -n $NS exec client -- wget -qO- -T 3 http://search 2>/dev/null)"; }
alias_ok && wait_for 30 works

check 2 "search is an ExternalName Service for the api Service" alias_ok
check 1 "DNS answers search with a CNAME to it" cname
check 2 "http://search reaches the search backend from client" works
