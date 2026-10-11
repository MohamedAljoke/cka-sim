# kubectl -n wl-pending describe pod -l app=api     -> 0/3 nodes ... didn't match Pod's node affinity/selector
# kubectl -n wl-pending describe pod -l app=batch   -> 0/3 nodes ... Insufficient memory
kubectl -n wl-pending patch deployment api --type=json -p='[{"op":"remove","path":"/spec/template/spec/affinity"}]'
kubectl -n wl-pending set resources deployment batch --requests=memory=128Mi
kubectl -n wl-pending rollout status deployment/api --timeout=120s
kubectl -n wl-pending rollout status deployment/batch --timeout=120s
