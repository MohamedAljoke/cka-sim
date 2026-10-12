kubectl -n kube-system get cm coredns -o yaml | sed 's/^\( *\)errors$/\1errors\n\1log/' | kubectl replace -f -
kubectl -n kube-system rollout restart deployment coredns
kubectl -n kube-system rollout status deployment coredns --timeout=120s
until kubectl -n kube-system logs -l k8s-app=kube-dns --tail=200 | grep -m1 'IN inventory.nw-coredns-log.svc.cluster.local' > /opt/course/nw-coredns-log/query.log; do
  kubectl -n nw-coredns-log exec client -- nslookup inventory.nw-coredns-log.svc.cluster.local >/dev/null 2>&1
  sleep 2
done
kubectl -n kube-system get cm coredns -o yaml | sed '/^ *log$/d' | kubectl replace -f -
kubectl -n kube-system rollout restart deployment coredns
kubectl -n kube-system rollout status deployment coredns --timeout=120s
