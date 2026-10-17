NS=nw-loadbalancer
svc() { kubectl -n $NS get service shop -o jsonpath="$1"; }
shape() { eq "LoadBalancer 80" "$(svc '{.spec.type} {.spec.ports[0].port}')"; }
pending() { grep -qi pending "$COURSE/external-ip.txt"; }
node_port() { [ -n "$(svc '{.spec.ports[0].nodePort}')" ] && eq "$(svc '{.spec.ports[0].nodePort}')" "$(tr -d ' \r\n' < "$COURSE/node-port.txt" 2>/dev/null)"; }
ip() { kubectl get node cka-sim-worker -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}'; }
reachable() { node_port && curl -s -m 3 "http://$(ip):$(svc '{.spec.ports[0].nodePort}')" | grep -q 'Welcome to nginx'; }
node_port && wait_for 20 reachable

check 2 "Service shop is a LoadBalancer on port 80" shape
check 1 "external-ip.txt says the external IP is pending" pending
check 1 "node-port.txt holds the Service's nodePort" node_port
check 1 "That nodePort serves nginx" reachable
