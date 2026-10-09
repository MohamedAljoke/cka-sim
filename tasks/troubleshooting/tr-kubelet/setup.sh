# Point the kubelet unit at a binary that does not exist, and stop it starting on boot.
on_host "sed -i 's|^ExecStart=/usr/bin/kubelet |ExecStart=/usr/local/bin/kubelet |' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf \
  && systemctl daemon-reload && systemctl disable kubelet 2>/dev/null; systemctl restart kubelet || true"
not_ready() { [ "$(k get node cka3962-worker -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')" != True ]; }
wait_for 120 not_ready
