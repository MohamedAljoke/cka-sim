manifest=/etc/kubernetes/manifests/etcd.yaml
grep -q '^    - --cert-file=/etc/kubernetes/pki/etcd/server.cert$' "$manifest" || exit 0
sed -i 's|^    - --cert-file=/etc/kubernetes/pki/etcd/server.cert$|    - --cert-file=/etc/kubernetes/pki/etcd/server.crt|' "$manifest"

api_up() { eq ok "$(kubectl get --raw /readyz --request-timeout=5s 2>/dev/null)"; }
wait_for 180 api_up
