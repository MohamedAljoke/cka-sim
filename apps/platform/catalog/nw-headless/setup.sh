fresh_ns nw-headless
fresh_course
kubectl -n nw-headless create deployment db --image=nginx:1.27-alpine --replicas=3 >/dev/null
kubectl -n nw-headless run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-headless rollout status deployment/db --timeout=120s >/dev/null
kubectl -n nw-headless wait --for=condition=Ready pod/client --timeout=120s >/dev/null
