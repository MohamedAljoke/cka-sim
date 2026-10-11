fresh_ns wl-scale
kubectl -n wl-scale create deployment web --image=nginx:1.27 --replicas=1
kubectl -n wl-scale rollout status deployment/web --timeout=120s
