manifest=/etc/kubernetes/manifests/kube-apiserver.yaml
api_up() { eq ok "$(kubectl get --raw /readyz --request-timeout=5s 2>/dev/null)"; }
nodes() { kubectl get nodes --request-timeout=5s; }
etcd_port() { grep -q '^    - --etcd-servers=https://127.0.0.1:2379$' "$manifest"; }
etcd_untouched() { api_up && grep -q '^    - --listen-client-urls=https://127.0.0.1:2379,' /etc/kubernetes/manifests/etcd.yaml; }
etcd_port && wait_for 180 api_up

check 2 "kube-apiserver uses etcd's client port 2379" etcd_port
check 3 "The API server is ready" api_up
check 1 "kubectl get nodes works" nodes
check 1 "etcd still listens on its usual client URLs" etcd_untouched
