fresh_ns tr-controller-manager
manifest=/etc/kubernetes/manifests/kube-controller-manager.yaml
sed -i 's|^    - --kubeconfig=/etc/kubernetes/controller-manager.config$|    - --kubeconfig=/etc/kubernetes/controller-manager.conf|' "$manifest"
sed -i 's|^    - --kubeconfig=/etc/kubernetes/controller-manager.conf$|    - --kubeconfig=/etc/kubernetes/controller-manager.config|' "$manifest"

running() { crictl ps -q --name '^kube-controller-manager$' | grep -q .; }
down() { ! running; }
wait_for 90 down

kubectl -n tr-controller-manager create deployment web --image=nginx:1.27-alpine --replicas=2 >/dev/null
