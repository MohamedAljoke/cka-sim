fresh_ns nw-dns-names
fresh_course
kubectl -n nw-dns-names create deployment api --image=nginx:1.27-alpine >/dev/null
kubectl -n nw-dns-names expose deployment api --port=80 >/dev/null
kubectl -n nw-dns-names run worker --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-dns-names wait --for=condition=Ready pod/worker --timeout=120s >/dev/null
