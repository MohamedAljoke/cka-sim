cfg=/var/lib/kubelet/config.yaml
endpoint() { sed -n 's/^containerRuntimeEndpoint: *//p' "$cfg"; }
containerd_endpoint() { [[ "$(endpoint)" =~ ^(unix://)?(/var)?/run/containerd/containerd\.sock$ ]]; }
node_ready() { eq True "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
containerd_endpoint && systemctl is-active --quiet kubelet && wait_for 90 node_ready

check 2 "The kubelet config points at the containerd socket" containerd_endpoint
check 2 "The kubelet service is active" systemctl is-active --quiet kubelet
check 3 "Node cka-sim-worker2 is Ready" node_ready
