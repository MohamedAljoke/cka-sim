NS=wl-init
init() { kubectl -n $NS get deployment web -o jsonpath="{.spec.template.spec.initContainers[?(@.name==\"wait-for-db\")].$1}"; }
init_ok() { eq busybox:1.36 "$(init image)" && [[ "$(init command)$(init args)" == *db* ]]; }
svc_ok() { eq 80 "$(kubectl -n $NS get service db -o jsonpath='{.spec.ports[0].port}')" && [ -n "$(kubectl -n $NS get endpointslices -l kubernetes.io/service-name=db -o jsonpath='{.items[*].endpoints[*].addresses[*]}')" ]; }
D() { kubectl -n $NS get deployment web -o jsonpath="$1"; }
ready() { init_ok && eq 2 "$(D '{.status.updatedReplicas}')" && eq 2 "$(D '{.status.readyReplicas}')"; }
init_ok && wait_for 120 ready

check 2 "web has an init container wait-for-db (busybox:1.36) that checks for db" init_ok
check 2 "Service db exists on port 80 with endpoints" svc_ok
check 2 "web has 2 Ready Pods with the init container" ready
