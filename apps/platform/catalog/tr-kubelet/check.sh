node_ready() { eq True "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
# A stopped kubelet never makes the node Ready, so only a running one is worth waiting for.
systemctl is-active --quiet kubelet && wait_for 90 node_ready

check 4 "node cka-sim-worker is Ready" node_ready
check 2 "kubelet service is active" systemctl is-active --quiet kubelet
check 2 "kubelet service is enabled at boot" systemctl is-enabled --quiet kubelet
