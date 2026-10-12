# kubectl -n tr-controller-manager get deploy,rs,pods -> web 0/2, no ReplicaSet: nothing reconciles Deployments
# kubectl -n kube-system get pods                      -> kube-controller-manager-... CrashLoopBackOff
# kubectl -n kube-system logs kube-controller-manager-cka-sim-control-plane -> stat /etc/kubernetes/controller-manager.config: no such file
sed -i 's|--kubeconfig=/etc/kubernetes/controller-manager.config|--kubeconfig=/etc/kubernetes/controller-manager.conf|' /etc/kubernetes/manifests/kube-controller-manager.yaml
