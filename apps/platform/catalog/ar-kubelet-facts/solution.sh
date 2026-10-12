# systemctl cat kubelet | grep -- --config -> KUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yaml
cfg=/var/lib/kubelet/config.yaml
{
  echo "config: $cfg"
  grep '^staticPodPath:' "$cfg"
  grep '^cgroupDriver:' "$cfg"
} > /opt/course/ar-kubelet-facts/kubelet.txt
