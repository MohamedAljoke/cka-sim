NS=nw-named-port
by_name() { eq "80 http" "$(kubectl -n $NS get service api -o jsonpath='{.spec.ports[0].port} {.spec.ports[0].targetPort}' 2>/dev/null)"; }
D() { kubectl -n $NS get deployment api -o jsonpath="$1"; }
moved() { eq "9090 http 9090" "$(D '{.spec.template.spec.containers[0].ports[0].containerPort} {.spec.template.spec.containers[0].ports[0].name} {.spec.template.spec.containers[0].env[?(@.name=="PORT")].value}')"; }
rolled() { moved && eq "2 2 2" "$(D '{.status.replicas} {.status.updatedReplicas} {.status.readyReplicas}')"; }
serves() { by_name && rolled && eq "api on 9090" "$(kubectl -n $NS exec client -- wget -qO- -T 3 http://api.$NS 2>/dev/null)"; }
moved && wait_for 90 serves

check 2 "Service api targets the port named http" by_name
check 1 "api listens on 9090, with container port http = 9090" moved
check 2 "http://api.nw-named-port serves the app on 9090" wait_for 30 serves
