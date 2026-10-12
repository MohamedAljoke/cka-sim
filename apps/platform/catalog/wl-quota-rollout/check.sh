NS=wl-quota-rollout
D() { kubectl -n $NS get deployment api -o jsonpath="$1"; }
reason_ok() { grep -q 'exceeded quota' "$COURSE/reason.txt" 2>/dev/null; }
done_rolling() { eq nginx:1.27-alpine "$(D '{.spec.template.spec.containers[0].image}')" && eq "2 2 2" "$(D '{.spec.replicas} {.status.updatedReplicas} {.status.readyReplicas}')" && [ -z "$(D '{.status.unavailableReplicas}')" ]; }
kept() { done_rolling && eq 400m "$(D '{.spec.template.spec.containers[0].resources.requests.cpu}')" && eq 1 "$(kubectl -n $NS get resourcequota team-cpu -o jsonpath='{.spec.hard.requests\.cpu}')"; }
[ "$(D '{.status.conditions[?(@.type=="Progressing")].reason}')" != ProgressDeadlineExceeded ] && wait_for 90 done_rolling

check 2 "reason.txt holds the quota error from the ReplicaSet's events" reason_ok
check 3 "api runs 2 Ready replicas of nginx:1.27-alpine" done_rolling
check 1 "The quota, replicas and CPU request are unchanged" kept
