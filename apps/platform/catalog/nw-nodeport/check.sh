NS=nw-nodeport
svc() { kubectl -n $NS get service web -o jsonpath="$1"; }
shape() { eq "NodePort 30080" "$(svc '{.spec.type} {.spec.ports[0].nodePort}')"; }
ip() { kubectl get node "$1" -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}'; }
reach() { curl -s -m 3 "http://$(ip "$1"):30080" | grep -q 'Welcome to nginx'; }
worker() { reach cka-sim-worker2; }
everywhere() { reach cka-sim-control-plane && reach cka-sim-worker && reach cka-sim-worker2; }
shape && wait_for 20 worker

check 2 "Service web is a NodePort Service on nodePort 30080" shape
check 2 "http://<cka-sim-worker2 IP>:30080 serves nginx" worker
check 1 "Port 30080 answers on every node" everywhere
