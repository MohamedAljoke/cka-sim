# ssh cka3962-worker, then sudo -i:
#   systemctl status kubelet        -> "activating (auto-restart)", exit code 203/EXEC
#   journalctl -u kubelet | tail    -> "Failed to locate executable /usr/local/bin/kubelet"
#   systemctl cat kubelet           -> shows the drop-in with the wrong ExecStart path
#   which kubelet                   -> /usr/bin/kubelet
on_host "sed -i 's|^ExecStart=/usr/local/bin/kubelet |ExecStart=/usr/bin/kubelet |' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf \
  && systemctl daemon-reload && systemctl enable --now kubelet && systemctl restart kubelet"
