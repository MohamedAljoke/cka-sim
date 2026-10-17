fresh_ns nw-pod-facts
fresh_course
kubectl -n nw-pod-facts run tracker --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-pod-facts wait --for=condition=Ready pod/tracker --timeout=120s >/dev/null
