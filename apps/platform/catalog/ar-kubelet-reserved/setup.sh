cfg=/var/lib/kubelet/config.yaml
clean=$(mktemp)
awk '/^systemReserved:/{skip=1; next} skip && /^  /{next} {skip=0} !/^  memory\.available:/' "$cfg" > "$clean"
if ! cmp -s "$cfg" "$clean"; then
  cat "$clean" > "$cfg"
  systemctl restart kubelet
fi
rm -f "$clean"
reset_done() { eq "$(kubectl get node "$(hostname)" -o jsonpath='{.status.capacity.cpu}')" "$(kubectl get node "$(hostname)" -o jsonpath='{.status.allocatable.cpu}')"; }
ready() { eq True "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
wait_for 120 ready
wait_for 60 reset_done
