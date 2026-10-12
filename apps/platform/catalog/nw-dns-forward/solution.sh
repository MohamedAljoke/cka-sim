ip=$(kubectl -n nw-dns-forward get service corp-dns -o jsonpath='{.spec.clusterIP}')
corefile=$(mktemp)
kubectl -n kube-system get configmap coredns -o jsonpath='{.data.Corefile}' > "$corefile"
cat >> "$corefile" <<EOT
corp.internal:53 {
    errors
    cache 30
    forward . $ip
}
EOT
kubectl -n kube-system create configmap coredns --from-file=Corefile="$corefile" --dry-run=client -o yaml | kubectl replace -f -
kubectl -n kube-system rollout restart deployment coredns
kubectl -n kube-system rollout status deployment coredns --timeout=120s
