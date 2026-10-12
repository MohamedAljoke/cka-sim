NS=nw-clusterip
svc() { kubectl -n $NS get service api -o jsonpath="$1"; }
shape() { eq "ClusterIP 80" "$(svc '{.spec.type} {.spec.ports[0].port}')"; }
serves() { eq "api ok" "$(kubectl -n $NS exec client -- wget -qO- -T 3 http://api.$NS 2>/dev/null)"; }
both() { serves && eq 2 "$(kubectl -n $NS get endpointslices -l kubernetes.io/service-name=api -o jsonpath='{.items[*].endpoints[*].addresses[*]}' | wc -w)"; }
shape && wait_for 20 serves

check 1 "Service api is a ClusterIP Service on port 80" shape
check 3 "http://api.nw-clusterip answers from inside the cluster" serves
check 1 "Both Pods of api are endpoints" both
