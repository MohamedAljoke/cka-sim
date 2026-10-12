NS=wl-rollout-pause
D() { kubectl -n $NS get deployment web -o jsonpath="$1"; }
changed() { eq nginx:1.27-alpine "$(D '{.spec.template.spec.containers[0].image}')" && eq "50m 32Mi" "$(D '{.spec.template.spec.containers[0].resources.requests.cpu} {.spec.template.spec.containers[0].resources.requests.memory}')"; }
done_rolling() { changed && [ "$(D '{.spec.paused}')" != true ] && eq 3 "$(D '{.status.updatedReplicas}')" && eq 3 "$(D '{.status.readyReplicas}')"; }
one_revision() { done_rolling && eq 2 "$(kubectl -n $NS get replicasets -o name | wc -l)"; }
cause() { changed && eq "upgrade to 1.27 with requests" "$(D '{.metadata.annotations.kubernetes\.io/change-cause}')"; }
changed && wait_for 90 done_rolling

check 2 "web runs nginx:1.27-alpine with requests 50m CPU and 32Mi memory" changed
check 2 "The rollout is resumed and finished" done_rolling
check 2 "Both changes came in one revision (2 ReplicaSets in total)" one_revision
check 1 "The change cause is recorded" cause
