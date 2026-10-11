# kubectl -n release rollout history deploy/shop               -> revisions 1, 2, 3
# kubectl -n release rollout history deploy/shop --revision=2  -> nginx:1.27-alpine, the last good one
kubectl -n release rollout undo deploy/shop
kubectl -n release rollout status deploy/shop --timeout=180s
kubectl -n release get deploy shop -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}' > "$COURSE/revision.txt"
kubectl -n release patch deploy shop -p '{"spec":{"strategy":{"rollingUpdate":{"maxUnavailable":1,"maxSurge":2}}}}'
