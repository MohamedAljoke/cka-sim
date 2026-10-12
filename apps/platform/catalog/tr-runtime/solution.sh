# systemctl status kubelet       -> activating (auto-restart)
# journalctl -u kubelet | tail   -> validate service connection: ... dial unix /var/run/cri-dockerd.sock: no such file or directory
# ls /run/containerd/            -> containerd.sock is the socket that exists
sed -i 's|^containerRuntimeEndpoint: .*|containerRuntimeEndpoint: unix:///run/containerd/containerd.sock|' /var/lib/kubelet/config.yaml
systemctl restart kubelet
