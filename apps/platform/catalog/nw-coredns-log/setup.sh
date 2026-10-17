fresh_ns nw-coredns-log
fresh_course
kubectl -n nw-coredns-log run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-coredns-log wait --for=condition=Ready pod/client --timeout=120s >/dev/null
