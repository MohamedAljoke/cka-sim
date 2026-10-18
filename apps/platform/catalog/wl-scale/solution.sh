kubectl -n wl-scale scale deployment web --replicas=4
kubectl -n wl-scale rollout status deployment/web --timeout=120s
