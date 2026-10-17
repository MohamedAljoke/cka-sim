NS=nw-headless
headless() { eq "None app=db" "$(kubectl -n $NS get service db -o jsonpath='{.spec.clusterIP} app={.spec.selector.app}' 2>/dev/null)"; }
pod_ips() { kubectl -n $NS get pods -l app=db -o jsonpath='{.items[*].status.podIP}' | tr ' ' '\n' | sort | xargs; }
written() { grep -oE '([0-9]{1,3}\.){3}[0-9]{1,3}' "$COURSE/ips.txt" 2>/dev/null | sort -u | xargs; }
resolves() { headless && eq "$(pod_ips)" "$(kubectl -n $NS exec client -- nslookup db.$NS.svc.cluster.local 2>/dev/null | sed -n '/^Name:/,$p' | grep -oE '([0-9]{1,3}\.){3}[0-9]{1,3}' | sort -u | xargs)"; }
headless && wait_for 30 resolves

check 2 "Service db is headless and selects app=db" headless
check 1 "Its DNS name returns the 3 Pod IPs" resolves
check 2 "ips.txt lists exactly the 3 Pod IPs" eq "$(pod_ips)" "$(written)"
