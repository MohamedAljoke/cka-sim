pki=/var/lib/kubelet/pki
trusted() { openssl verify -CAfile /etc/kubernetes/pki/ca.crt "$1" >/dev/null 2>&1; }
trusted "$pki/kubelet-client-current.pem" && systemctl is-active -q kubelet && exit 0
for f in "$pki"/kubelet-client-2*.pem; do trusted "$f" && good=$f; done
ln -sfn "$good" "$pki/kubelet-client-current.pem"
systemctl restart kubelet

ready() { eq True "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
wait_for 120 ready
