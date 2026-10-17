fresh_ns wl-pdb
kubectl -n wl-pdb create deployment api --image=nginx:1.27-alpine --replicas=3 >/dev/null
kubectl -n wl-pdb rollout status deployment/api --timeout=120s >/dev/null
