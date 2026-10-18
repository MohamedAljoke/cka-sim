# kubectl get nodes                 -> connection refused: the API server isn't listening
# crictl ps -a --name kube-apiserver -> Exited, restarting
# crictl logs <id> | tail           -> dial tcp 127.0.0.1:2389: connect: connection refused (etcd)
# grep listen-client /etc/kubernetes/manifests/etcd.yaml -> etcd listens on 2379
sed -i 's|--etcd-servers=https://127.0.0.1:2389|--etcd-servers=https://127.0.0.1:2379|' /etc/kubernetes/manifests/kube-apiserver.yaml
