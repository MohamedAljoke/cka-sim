cm() { kubectl -n kube-system get configmap coredns -o yaml; }
cm | sed 's/^\( *\)kubernetes cluster.local /\1kubernetess cluster.local /' | kubectl replace -f - >/dev/null
old=$(kubectl -n kube-system get pods -l k8s-app=kube-dns -o name)
kubectl -n kube-system delete pod -l k8s-app=kube-dns --wait=false >/dev/null

fresh_ns tr-coredns
kubectl -n tr-coredns run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n tr-coredns wait --for=condition=Ready pod/client --timeout=120s >/dev/null
crashed() { kubectl -n kube-system get pods -l k8s-app=kube-dns -o jsonpath='{.items[*].status.containerStatuses[0].restartCount}' | grep -q '[1-9]'; }
wait_for 120 crashed
kubectl -n kube-system wait --for=delete $old --timeout=90s >/dev/null 2>&1
