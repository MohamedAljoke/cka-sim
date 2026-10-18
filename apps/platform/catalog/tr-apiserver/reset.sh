manifest=/etc/kubernetes/manifests/kube-apiserver.yaml
grep -q '^    - --etcd-servers=https://127.0.0.1:2389$' "$manifest" || exit 0
sed -i 's|^    - --etcd-servers=https://127.0.0.1:2389$|    - --etcd-servers=https://127.0.0.1:2379|' "$manifest"

api_up() { eq ok "$(kubectl get --raw /readyz --request-timeout=5s 2>/dev/null)"; }
wait_for 180 api_up
