# kubectl -n nw-cni-bindir describe pod -l app=web -> failed to find plugin "ptp" in path [/opt/cni/plugins]
# ls /opt/cni/bin                                  -> the plugins are here
# grep -n -A2 cni /etc/containerd/config.toml       -> bin_dir = "/opt/cni/plugins"
sed -i 's|bin_dir = "/opt/cni/plugins"|bin_dir = "/opt/cni/bin"|' /etc/containerd/config.toml
systemctl restart containerd
