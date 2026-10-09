dep() { kubectl -n release get deploy shop -o jsonpath="$1"; }
healthy() { eq 3 "$(dep '{.status.updatedReplicas}')" && eq 3 "$(dep '{.status.availableReplicas}')"; }
on_good_image() { eq nginx:1.27-alpine "$(dep '{.spec.template.spec.containers[0].image}')"; }
wait_for 60 healthy

check 2 "shop runs nginx:1.27-alpine again" on_good_image
rolled_back_and_healthy() { on_good_image && healthy; }
check 1 "all 3 replicas are updated and available" rolled_back_and_healthy
check 2 "revision.txt holds the current revision" eq "$(dep '{.metadata.annotations.deployment\.kubernetes\.io/revision}')" "$(tr -d '[:space:]' < "$COURSE/revision.txt" 2>/dev/null)"
check 1 "maxUnavailable is 1" eq 1 "$(dep '{.spec.strategy.rollingUpdate.maxUnavailable}')"
check 1 "maxSurge is 2" eq 2 "$(dep '{.spec.strategy.rollingUpdate.maxSurge}')"
