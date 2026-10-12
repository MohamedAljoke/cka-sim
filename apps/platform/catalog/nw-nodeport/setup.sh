fresh_ns nw-nodeport
kubectl -n nw-nodeport create deployment web --image=nginx:1.27-alpine --replicas=2 >/dev/null
kubectl -n nw-nodeport rollout status deployment/web --timeout=120s >/dev/null
