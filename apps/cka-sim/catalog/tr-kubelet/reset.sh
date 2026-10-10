dropin=/etc/systemd/system/kubelet.service.d/10-kubeadm.conf
if grep -q '^ExecStart=/usr/bin/kubelet ' "$dropin" && systemctl is-enabled -q kubelet && systemctl is-active -q kubelet; then
  exit 0
fi
sed -i 's|^ExecStart=/usr/local/bin/kubelet |ExecStart=/usr/bin/kubelet |' "$dropin"
systemctl daemon-reload
systemctl enable --now kubelet
systemctl restart kubelet

ready() { [ "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')" = True ]; }
wait_for 120 ready
