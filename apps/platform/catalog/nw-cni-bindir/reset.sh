cfg=/etc/containerd/config.toml
grep -q '^\[plugins."io.containerd.grpc.v1.cri".cni\]$' "$cfg" || exit 0
sed -i '/^\[plugins."io.containerd.grpc.v1.cri".cni\]$/,+2d' "$cfg"
systemctl restart containerd
up() { crictl info >/dev/null; }
wait_for 60 up
