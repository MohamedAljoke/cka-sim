dep() { k -n release get deploy shop -o jsonpath="$1"; }
healthy() { eq 3 "$(dep '{.status.updatedReplicas}')" && eq 3 "$(dep '{.status.availableReplicas}')"; }
wait_for 60 healthy
check 2 "shop runs nginx:1.27-alpine again" eq nginx:1.27-alpine "$(dep '{.spec.template.spec.containers[0].image}')"
check 1 "all 3 replicas are updated and available" healthy
check 2 "revision.txt holds the current revision" eq "$(dep '{.metadata.annotations.deployment\.kubernetes\.io/revision}')" "$(on_host "cat $(course_dir)/revision.txt" | tr -d '[:space:]')"
check 1 "maxUnavailable is 1" eq 1 "$(dep '{.spec.strategy.rollingUpdate.maxUnavailable}')"
check 1 "maxSurge is 2" eq 2 "$(dep '{.spec.strategy.rollingUpdate.maxSurge}')"
