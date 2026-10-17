fresh_ns wl-init
kubectl -n wl-init create deployment db --image=nginx:1.27-alpine >/dev/null
kubectl -n wl-init create deployment web --image=nginx:1.27-alpine --replicas=2 >/dev/null
kubectl -n wl-init rollout status deployment/db --timeout=120s >/dev/null
kubectl -n wl-init rollout status deployment/web --timeout=120s >/dev/null
