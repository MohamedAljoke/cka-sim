fresh_course
KC="$COURSE/dev.kubeconfig"
src=/etc/kubernetes/admin.conf
kubectl --kubeconfig "$KC" config set-cluster cka-sim --server=https://cka-sim-control-plane:6444 >/dev/null
kubectl --kubeconfig "$KC" config set clusters.cka-sim.certificate-authority-data "$(kubectl --kubeconfig $src config view --raw --minify -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')" >/dev/null
kubectl --kubeconfig "$KC" config set users.admin.client-certificate-data "$(kubectl --kubeconfig $src config view --raw --minify -o jsonpath='{.users[0].user.client-certificate-data}')" >/dev/null
kubectl --kubeconfig "$KC" config set users.admin.client-key-data "$(kubectl --kubeconfig $src config view --raw --minify -o jsonpath='{.users[0].user.client-key-data}')" >/dev/null
kubectl --kubeconfig "$KC" config set-context dev --cluster=cka-sim --user=dev-admin >/dev/null
chmod 0644 "$KC"
