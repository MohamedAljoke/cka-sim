corefile() { kubectl -n kube-system get configmap coredns -o jsonpath='{.data.Corefile}'; }
logged() { grep -q 'inventory\.nw-coredns-log\.svc\.cluster\.local' "$COURSE/query.log" 2>/dev/null && grep -qE '"(A|AAAA) IN ' "$COURSE/query.log"; }
log_off() { logged && ! corefile | grep -qE '^[[:space:]]*log[[:space:]]*$'; }
intact() { log_off && corefile | grep -q 'kubernetes cluster.local' && corefile | grep -q 'forward . /etc/resolv.conf' && eq 2 "$(kubectl -n kube-system get deployment coredns -o jsonpath='{.status.readyReplicas}')"; }

check 3 "query.log holds CoreDNS's log line for the inventory lookup" logged
check 1 "Query logging is off again" log_off
check 1 "The rest of the Corefile is intact and CoreDNS is Ready" intact
