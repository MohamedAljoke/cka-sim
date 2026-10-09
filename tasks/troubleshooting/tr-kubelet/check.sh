node_ready() { eq True "$(k get node cka3962-worker -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
wait_for 90 node_ready
check 4 "node cka3962-worker is Ready" node_ready
check 2 "kubelet service is active" on_host "systemctl is-active --quiet kubelet"
check 2 "kubelet service is enabled at boot" on_host "systemctl is-enabled --quiet kubelet"
