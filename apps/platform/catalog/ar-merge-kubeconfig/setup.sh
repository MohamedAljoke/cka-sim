fresh_ns ar-merge-b
fresh_course
kubectl -n ar-merge-b create configmap marker --from-literal=team=b >/dev/null
src=/etc/kubernetes/admin.conf
mk() {
  local f="$COURSE/$1.kubeconfig"
  kubectl --kubeconfig "$f" config set-cluster "$2" --server=https://cka-sim-control-plane:6443 >/dev/null
  kubectl --kubeconfig "$f" config set "clusters.$2.certificate-authority-data" "$(kubectl --kubeconfig $src config view --raw --minify -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')" >/dev/null
  kubectl --kubeconfig "$f" config set "users.$3.client-certificate-data" "$(kubectl --kubeconfig $src config view --raw --minify -o jsonpath='{.users[0].user.client-certificate-data}')" >/dev/null
  kubectl --kubeconfig "$f" config set "users.$3.client-key-data" "$(kubectl --kubeconfig $src config view --raw --minify -o jsonpath='{.users[0].user.client-key-data}')" >/dev/null
  kubectl --kubeconfig "$f" config set-context "$1" --cluster="$2" --user="$3" >/dev/null
  kubectl --kubeconfig "$f" config use-context "$1" >/dev/null
  chmod 0644 "$f"
}
mk team-a cluster-a admin-a
mk team-b cluster-b admin-b
