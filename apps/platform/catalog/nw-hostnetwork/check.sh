NS=nw-hostnetwork
P() { kubectl -n $NS get pod node-agent -o jsonpath="$1" 2>/dev/null; }
node_ip() { kubectl get node cka-sim-worker2 -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}'; }
host_net() { eq "true cka-sim-worker2 Running" "$(P '{.spec.hostNetwork} {.spec.nodeName} {.status.phase}')"; }
same_ip() { host_net && eq "$(node_ip)" "$(P '{.status.podIP}')"; }
answers() { host_net && eq agent "$(curl -s -m 3 "http://$(node_ip):18080")"; }
no_service() { answers && [ -z "$(kubectl -n $NS get services -o name)" ]; }
P "{.metadata.name}" >/dev/null && wait_for 60 host_net

check 2 "node-agent runs on cka-sim-worker2 with hostNetwork" host_net
check 1 "Its Pod IP is the node's IP" same_ip
check 2 "<cka-sim-worker2 IP>:18080 answers agent" answers
check 1 "No Service was needed" no_service
