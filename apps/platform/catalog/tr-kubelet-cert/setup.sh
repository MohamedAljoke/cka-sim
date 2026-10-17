pki=/var/lib/kubelet/pki
trusted() { openssl verify -CAfile /etc/kubernetes/pki/ca.crt "$1" >/dev/null 2>&1; }
for f in "$pki"/kubelet-client-2*.pem; do trusted "$f" || rm -f "$f"; done
key=$(mktemp)
crt=$(mktemp)
openssl req -x509 -newkey rsa:2048 -nodes -days 365 -keyout "$key" -out "$crt" \
  -subj "/O=system:nodes/CN=system:node:$(hostname)" >/dev/null 2>&1
bad="$pki/kubelet-client-$(date +%Y-%m-%d-%H-%M-%S).pem"
cat "$crt" "$key" > "$bad"
chmod 600 "$bad"
rm -f "$key" "$crt"
ln -sfn "$bad" "$pki/kubelet-client-current.pem"
systemctl restart kubelet

not_ready() { [ "$(kubectl get node "$(hostname)" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')" != True ]; }
wait_for 150 not_ready
