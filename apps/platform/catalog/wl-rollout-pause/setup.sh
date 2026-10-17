fresh_ns wl-rollout-pause
kubectl -n wl-rollout-pause create deployment web --image=nginx:1.26-alpine --replicas=3 >/dev/null
kubectl -n wl-rollout-pause rollout status deployment/web --timeout=120s >/dev/null
