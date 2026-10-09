# systemctl status kubelet        -> "activating (auto-restart)", exit code 203/EXEC
# journalctl -u kubelet | tail    -> "Failed to locate executable /usr/local/bin/kubelet"
# systemctl cat kubelet           -> the drop-in has the wrong ExecStart path
# which kubelet                   -> /usr/bin/kubelet
sed -i 's|^ExecStart=/usr/local/bin/kubelet |ExecStart=/usr/bin/kubelet |' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
systemctl daemon-reload
systemctl enable --now kubelet
systemctl restart kubelet
