cfg=/var/lib/kubelet/config.yaml
if grep -qE '^containerLogMax(Size|Files):' "$cfg"; then
  sed -i -E '/^containerLogMax(Size|Files):/d' "$cfg"
  systemctl restart kubelet
fi
ready() { eq True "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
wait_for 120 ready
