NS=wl-canary
C() { kubectl -n $NS get deployment web-canary -o jsonpath="$1" 2>/dev/null; }
canary() { eq "1 1 web canary" "$(C '{.spec.replicas} {.status.readyReplicas} {.spec.template.metadata.labels.app} {.spec.template.metadata.labels.track}')"; }
endpoints() { kubectl -n $NS get endpointslices -l kubernetes.io/service-name=web -o jsonpath='{.items[*].endpoints[?(@.conditions.ready==true)].addresses[*]}' | wc -w; }
in_service() { canary && eq 4 "$(endpoints)"; }
hits() { for _ in $(seq 1 24); do kubectl -n $NS exec client -- wget -qO- -T 2 http://web 2>/dev/null; done | sort -u | xargs; }
both() { in_service && eq "canary stable" "$(hits)"; }
untouched() { in_service && eq "web" "$(kubectl -n $NS get service web -o jsonpath='{.spec.selector.app}')" && [ -z "$(kubectl -n $NS get service web -o jsonpath='{.spec.selector.track}')" ] && eq 1 "$(kubectl -n $NS get deployment web-stable -o jsonpath='{.metadata.generation}')"; }
C '{.metadata.name}' >/dev/null && wait_for 60 in_service

check 2 "web-canary runs 1 Ready Pod labelled app=web, track=canary" canary
check 2 "Service web has 4 Ready endpoints, the canary included" in_service
check 1 "Requests to web reach both stable and canary" both
check 1 "Service web and web-stable are unchanged" untouched
