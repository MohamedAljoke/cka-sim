NS=nw-readiness
endpoints() { [ -n "$(kubectl -n $NS get endpointslices -l kubernetes.io/service-name=web -o jsonpath='{.items[*].endpoints[?(@.conditions.ready==true)].addresses[*]}')" ]; }
serves() { [[ "$(kubectl -n $NS exec client -- wget -qO- -T 3 http://web.$NS 2>/dev/null)" == *"Welcome to nginx"* ]]; }
probe() { kubectl -n $NS get deployment web -o jsonpath='{.spec.template.spec.containers[0].readinessProbe.httpGet.port}'; }
has_probe() { serves && [ -n "$(probe)" ]; }
service() { kubectl -n $NS get service web -o jsonpath='{.spec.selector.app} {.spec.ports[0].port} {.spec.ports[0].targetPort}'; }
service_untouched() { serves && eq "web 80 80" "$(service)"; }
wait_for 60 endpoints && wait_for 20 serves

check 2 "Service web has Ready endpoints" endpoints
check 3 "http://web.nw-readiness serves the nginx page" serves
check 1 "The container still has an HTTP readiness probe" has_probe
check 1 "Service web was not modified" service_untouched
