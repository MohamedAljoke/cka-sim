current=/var/lib/kubelet/pki/kubelet-client-current.pem
trusted() { openssl verify -CAfile /etc/kubernetes/pki/ca.crt "$current"; }
node_ready() { eq True "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
trusted >/dev/null 2>&1 && wait_for 90 node_ready

check 3 "The kubelet's client certificate is signed by the cluster CA" trusted
check 3 "Node cka-sim-worker is Ready" node_ready
