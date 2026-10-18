# kubectl -n kube-system get pods -l k8s-app=kube-dns -> CrashLoopBackOff
# kubectl -n kube-system logs -l k8s-app=kube-dns      -> Unknown directive 'kubernetess'
kubectl -n kube-system get configmap coredns -o yaml | sed 's/kubernetess cluster.local/kubernetes cluster.local/' | kubectl replace -f -
kubectl -n kube-system rollout restart deployment coredns
kubectl -n kube-system rollout status deployment coredns --timeout=120s
