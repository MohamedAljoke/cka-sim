manifest=/etc/kubernetes/manifests/kube-apiserver.yaml
sed -i 's|^    - --etcd-servers=https://127.0.0.1:2389$|    - --etcd-servers=https://127.0.0.1:2379|' "$manifest"
sed -i 's|^    - --etcd-servers=https://127.0.0.1:2379$|    - --etcd-servers=https://127.0.0.1:2389|' "$manifest"

api_down() { ! kubectl get --raw /readyz --request-timeout=3s; }
wait_for 180 api_down
