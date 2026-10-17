cfg=/etc/containerd/config.toml
if ! grep -q '^\[plugins."io.containerd.grpc.v1.cri".cni\]$' "$cfg"; then
  printf '\n[plugins."io.containerd.grpc.v1.cri".cni]\n  bin_dir = "/opt/cni/plugins"\n  conf_dir = "/etc/cni/net.d"\n' >> "$cfg"
  systemctl restart containerd
fi
up() { crictl info >/dev/null; }
wait_for 60 up
fresh_ns nw-cni-bindir
kubectl -n nw-cni-bindir create deployment web --image=nginx:1.27-alpine --dry-run=client -o yaml |
  sed 's|^    spec:$|    spec:\n      nodeSelector:\n        kubernetes.io/hostname: cka-sim-worker2|' | kubectl apply -f - >/dev/null
sandbox_failed() { kubectl -n nw-cni-bindir get events --field-selector reason=FailedCreatePodSandBox -o name | grep -q .; }
wait_for 90 sandbox_failed
