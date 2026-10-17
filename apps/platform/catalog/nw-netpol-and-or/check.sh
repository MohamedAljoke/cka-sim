NS=nw-netpol-and-or
API=$(pod_ip $NS api)
monitoring() { connects nw-netpol-mon prometheus "$API" 8080 && ! connects nw-netpol-mon grafana "$API" 8080; }
local_blocked() { ! connects $NS prometheus "$API" 8080; }
entries() { kubectl -n $NS get networkpolicy api-from-monitoring -o jsonpath='{range .spec.ingress[0].from[*]}x{end}'; }
fixed() { eq api-from-monitoring "$(kubectl -n $NS get networkpolicy -o jsonpath='{.items[*].metadata.name}')" && eq x "$(entries)"; }

check 2 "prometheus in nw-netpol-mon reaches api, grafana is blocked" monitoring
check 2 "prometheus in nw-netpol-and-or is blocked" local_blocked
check 1 "api-from-monitoring is the only policy and has one from entry" fixed
