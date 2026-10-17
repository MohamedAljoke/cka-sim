fresh_ns nw-coredns-hosts
kubectl -n nw-coredns-hosts run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-coredns-hosts wait --for=condition=Ready pod/client --timeout=120s >/dev/null
