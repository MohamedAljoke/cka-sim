cm() { kubectl -n kube-system get configmap coredns -o yaml; }
cm | grep -q '^ *kubernetess cluster.local ' || exit 0
cm | sed 's/^\( *\)kubernetess cluster.local /\1kubernetes cluster.local /' | kubectl replace -f - >/dev/null
kubectl -n kube-system delete pod -l k8s-app=kube-dns --wait=false >/dev/null

ready() { eq 2 "$(kubectl -n kube-system get deployment coredns -o jsonpath='{.status.readyReplicas}')"; }
wait_for 120 ready
