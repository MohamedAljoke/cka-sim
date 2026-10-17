dropin=/etc/systemd/system/kubelet.service.d/10-kubeadm.conf
sed -i 's|^ExecStart=/usr/local/bin/kubelet |ExecStart=/usr/bin/kubelet |' "$dropin"
sed -i 's|^ExecStart=/usr/bin/kubelet |ExecStart=/usr/local/bin/kubelet |' "$dropin"
systemctl daemon-reload
systemctl disable kubelet 2>/dev/null
systemctl restart kubelet 2>/dev/null || true

not_ready() { [ "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')" != True ]; }
wait_for 120 not_ready
