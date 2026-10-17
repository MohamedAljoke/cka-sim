NS=nw-hostport
P() { kubectl -n $NS get pod web -o jsonpath="$1" 2>/dev/null; }
node_ip() { kubectl get node cka-sim-worker2 -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}'; }
shape() { eq "8081 80 cka-sim-worker2" "$(P '{.spec.containers[0].ports[0].hostPort} {.spec.containers[0].ports[0].containerPort} {.spec.nodeName}')" && [ "$(P '{.spec.hostNetwork}')" != true ]; }
serves() { shape && curl -s -m 3 "http://$(node_ip):8081" | grep -q 'Welcome to nginx'; }
own_ip() { serves && [ "$(P '{.status.podIP}')" != "$(node_ip)" ] && [ -z "$(kubectl -n $NS get services -o name)" ]; }
shape && wait_for 60 serves

check 2 "web maps hostPort 8081 to containerPort 80 on cka-sim-worker2" shape
check 2 "<cka-sim-worker2 IP>:8081 serves nginx" serves
check 1 "web keeps its own Pod IP, and there is no Service" own_ip
