fresh_course
dir=$(mktemp -d)
openssl genrsa -out "$dir/k" 2048 2>/dev/null
openssl req -new -key "$dir/k" -subj "/CN=ext-auditor/O=auditors/O=readonly" -out "$dir/r" 2>/dev/null
openssl x509 -req -in "$dir/r" -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -days 75 -out "$dir/c" 2>/dev/null
KC="$COURSE/auditor.kubeconfig"
kubectl --kubeconfig "$KC" config set-cluster cka-sim --server=https://cka-sim-control-plane:6443 --certificate-authority=/etc/kubernetes/pki/ca.crt --embed-certs >/dev/null
kubectl --kubeconfig "$KC" config set-credentials auditor --client-certificate="$dir/c" --client-key="$dir/k" --embed-certs >/dev/null
kubectl --kubeconfig "$KC" config set-context auditor --cluster=cka-sim --user=auditor >/dev/null
kubectl --kubeconfig "$KC" config use-context auditor >/dev/null
chmod 0644 "$KC"
rm -rf "$dir"
