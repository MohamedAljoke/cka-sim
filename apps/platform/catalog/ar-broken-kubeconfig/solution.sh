KC=/opt/course/ar-broken-kubeconfig/dev.kubeconfig
# kubectl --kubeconfig $KC config view -> server :6444 (the API server listens on 6443), context dev uses user dev-admin, which doesn't exist (the user is "admin"), no current-context
kubectl --kubeconfig $KC config set-cluster cka-sim --server=https://cka-sim-control-plane:6443
kubectl --kubeconfig $KC config set-context dev --user=admin
kubectl --kubeconfig $KC config use-context dev
kubectl --kubeconfig $KC get nodes
