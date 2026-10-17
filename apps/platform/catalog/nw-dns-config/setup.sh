fresh_ns nw-dns-config
fresh_course
kubectl -n nw-dns-config run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-dns-config wait --for=condition=Ready pod/client --timeout=120s >/dev/null
