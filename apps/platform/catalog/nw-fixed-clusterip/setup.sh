fresh_ns nw-fixed-clusterip
fresh_course
kubectl -n nw-fixed-clusterip create deployment cache --image=nginx:1.27-alpine >/dev/null
kubectl -n nw-fixed-clusterip run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-fixed-clusterip rollout status deployment/cache --timeout=120s >/dev/null
kubectl -n nw-fixed-clusterip wait --for=condition=Ready pod/client --timeout=120s >/dev/null
