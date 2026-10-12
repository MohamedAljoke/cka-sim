NS=nw-etp-local
ip() { kubectl get node "$1" -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}'; }
reach() { curl -s -m 3 "http://$(ip "$1"):30090" | grep -q 'Welcome to nginx'; }
listed() { tr -d ' \r' < "$COURSE/answering.txt" 2>/dev/null | grep . | sort | xargs; }
everywhere() { reach cka-sim-control-plane && reach cka-sim-worker && reach cka-sim-worker2; }
policy() { kubectl -n $NS get service web -o jsonpath='{.spec.externalTrafficPolicy}'; }
kept() { everywhere && eq "1 30090" "$(kubectl -n $NS get deployment web -o jsonpath='{.spec.replicas}') $(kubectl -n $NS get service web -o jsonpath='{.spec.ports[0].nodePort}')"; }
[ "$(policy)" = Cluster ] && wait_for 30 everywhere

check 2 "answering.txt names only cka-sim-worker2" eq cka-sim-worker2 "$(listed)"
check 3 "Port 30090 answers on every node" everywhere
check 1 "1 replica and nodePort 30090 are kept" kept
