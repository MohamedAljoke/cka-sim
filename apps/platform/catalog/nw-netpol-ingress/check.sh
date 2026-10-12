NS=nw-netpol-ingress
API=$(pod_ip $NS api)
frontend() { connects $NS frontend "$API" 8080 && ! connects $NS worker "$API" 8080; }
monitor() { connects nw-netpol-monitor probe "$API" 8080 && ! connects nw-netpol-outside stranger "$API" 8080; }
one_port() { connects $NS frontend "$API" 8080 && ! connects $NS frontend "$API" 9090; }
selects_api() { eq api "$(kubectl -n $NS get networkpolicy api-allow -o jsonpath='{.spec.podSelector.matchLabels.app}')"; }

check 2 "frontend reaches api:8080, worker is blocked" frontend
check 2 "probe in nw-netpol-monitor reaches api:8080, a role=frontend Pod in nw-netpol-outside is blocked" monitor
check 1 "Port 9090 is blocked even for frontend" one_port
check 1 "NetworkPolicy api-allow selects app=api" selects_api
