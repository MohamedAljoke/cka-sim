kubectl -n kube-system get configmap coredns -o jsonpath='{.data.Corefile}' | grep -q 'corp\.internal' || exit 0
corefile=$(mktemp)
cat > "$corefile" <<'COREFILE'
.:53 {
    errors
    health {
       lameduck 5s
    }
    ready
    kubernetes cluster.local in-addr.arpa ip6.arpa {
       pods insecure
       fallthrough in-addr.arpa ip6.arpa
       ttl 30
    }
    prometheus :9153
    forward . /etc/resolv.conf {
       max_concurrent 1000
    }
    cache 30 {
       disable success cluster.local
       disable denial cluster.local
    }
    loop
    reload
    loadbalance
}
COREFILE
kubectl -n kube-system create configmap coredns --from-file=Corefile="$corefile" --dry-run=client -o yaml | kubectl replace -f - >/dev/null
rm -f "$corefile"
kubectl -n kube-system rollout restart deployment coredns >/dev/null
kubectl -n kube-system rollout status deployment coredns --timeout=120s >/dev/null
