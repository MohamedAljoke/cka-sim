fresh_ns wl-graceful
kubectl -n wl-graceful create deployment web --image=nginx:1.27-alpine --replicas=2 >/dev/null
kubectl -n wl-graceful rollout status deployment/web --timeout=120s >/dev/null
