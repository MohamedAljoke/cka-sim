manifest=/etc/kubernetes/manifests/kube-controller-manager.yaml
grep -q '^    - --kubeconfig=/etc/kubernetes/controller-manager.config$' "$manifest" || exit 0
sed -i 's|^    - --kubeconfig=/etc/kubernetes/controller-manager.config$|    - --kubeconfig=/etc/kubernetes/controller-manager.conf|' "$manifest"

running() { crictl ps -q --name '^kube-controller-manager$' | grep -q .; }
wait_for 90 running
