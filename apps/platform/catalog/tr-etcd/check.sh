manifest=/etc/kubernetes/manifests/etcd.yaml
api_up() { eq ok "$(kubectl get --raw /readyz --request-timeout=5s 2>/dev/null)"; }
cert_ok() { grep -q '^    - --cert-file=/etc/kubernetes/pki/etcd/server.crt$' "$manifest"; }
logged() { grep -q 'server\.cert: no such file' "$COURSE/error.log"; }
cert_ok && wait_for 180 api_up

check 2 "error.log holds etcd's fatal error about its certificate file" logged
check 2 "etcd uses its server certificate /etc/kubernetes/pki/etcd/server.crt" cert_ok
check 3 "The API server is ready again" api_up
