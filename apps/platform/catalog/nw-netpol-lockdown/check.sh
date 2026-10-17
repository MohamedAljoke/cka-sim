NS=nw-netpol-lockdown
WEB=$(pod_ip $NS web)
resolves() { kubectl -n $NS exec app -- nslookup kubernetes.default.svc.cluster.local; }
inside() { ! connects $NS app "$WEB" 8080; }
from_outside() { ! connects nw-netpol-visitor visitor "$WEB" 8080; }
dns_only() { resolves && ! connects $NS app kubernetes.default.svc.cluster.local 443; }
types() { kubectl -n $NS get networkpolicy default-deny -o jsonpath='{.spec.policyTypes[*]}' | tr ' ' '\n' | sort | xargs; }

check 2 "app can't reach web inside the Namespace" inside
check 1 "visitor can't reach web:8080 from another Namespace" from_outside
check 2 "app resolves names but can't reach the API server on 443" dns_only
check 1 "default-deny lists both Ingress and Egress" eq "Egress Ingress" "$(types)"
