kubectl -n kube-system get cm coredns -o yaml |
  sed 's/^\( *\)ready$/\1ready\n\1hosts {\n\1   10.10.10.10 legacy-db.corp\n\1   fallthrough\n\1}/' | kubectl replace -f -
kubectl -n kube-system rollout restart deployment coredns
kubectl -n kube-system rollout status deployment coredns --timeout=120s
