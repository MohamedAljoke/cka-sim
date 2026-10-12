digest=$(cat /etc/cka-sim/wl-image-digest 2>/dev/null)
container() { kubectl -n wl-image-digest get deployment web -o jsonpath="{.spec.template.spec.containers[0].$1}"; }
pinned() { [ -n "$digest" ] && eq "cka-sim-registry:5000/team/web@$digest" "$(container image)"; }
policy() { eq IfNotPresent "$(container imagePullPolicy)"; }
answer() { tr -d '[:space:]' < "$COURSE/digest.txt" 2>/dev/null; }
rolled_out() { eq "2 2" "$(kubectl -n wl-image-digest get deployment web -o jsonpath='{.status.updatedReplicas} {.status.readyReplicas}')"; }

check 3 "web references the running image by its digest and runs 2 Pods of it" eval 'pinned && wait_for 90 rolled_out'
check 1 "web's imagePullPolicy is IfNotPresent" policy
check 1 "digest.txt holds the digest" eq "$digest" "$(answer)"
