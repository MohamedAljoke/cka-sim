# k -n release rollout history deploy/shop                 -> revisions 1, 2, 3
# k -n release rollout history deploy/shop --revision=2    -> nginx:1.27-alpine (the last good one)
k -n release rollout undo deploy/shop
k -n release rollout status deploy/shop --timeout=180s
rev=$(k -n release get deploy shop -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')
on_host "echo $rev > $(course_dir)/revision.txt"
k -n release patch deploy shop -p '{"spec":{"strategy":{"rollingUpdate":{"maxUnavailable":1,"maxSurge":2}}}}'
