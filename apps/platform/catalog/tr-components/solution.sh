# systemctl status kubelet                   -> a systemd service
# ls /etc/kubernetes/manifests/              -> etcd, kube-apiserver, kube-controller-manager, kube-scheduler
# kubectl -n kube-system get deploy,ds       -> coredns Deployment, kube-proxy DaemonSet
cat > /opt/course/tr-components/components.txt <<TXT
kubelet: systemd
kube-apiserver: static-pod
kube-scheduler: static-pod
kube-controller-manager: static-pod
etcd: static-pod
coredns: deployment
kube-proxy: daemonset
TXT
