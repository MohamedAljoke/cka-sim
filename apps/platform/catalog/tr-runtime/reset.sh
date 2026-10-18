cfg=/var/lib/kubelet/config.yaml
grep -q '^containerRuntimeEndpoint: unix:///run/containerd/containerd.sock$' "$cfg" && systemctl is-active -q kubelet && exit 0
sed -i 's|^containerRuntimeEndpoint: .*|containerRuntimeEndpoint: unix:///run/containerd/containerd.sock|' "$cfg"
systemctl restart kubelet

ready() { eq True "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
wait_for 120 ready
