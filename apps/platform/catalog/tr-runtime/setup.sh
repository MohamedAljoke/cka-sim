cfg=/var/lib/kubelet/config.yaml
sed -i 's|^containerRuntimeEndpoint: .*|containerRuntimeEndpoint: unix:///var/run/cri-dockerd.sock|' "$cfg"
systemctl restart kubelet 2>/dev/null || true

not_ready() { [ "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')" != True ]; }
wait_for 150 not_ready
