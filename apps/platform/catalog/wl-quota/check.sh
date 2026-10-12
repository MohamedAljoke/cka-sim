NS=wl-quota
lr() { kubectl -n $NS get limitrange defaults -o jsonpath="{.spec.limits[?(@.type==\"Container\")].$1}"; }
defaults_ok() {
  eq 100m "$(lr 'defaultRequest.cpu')" && eq 64Mi "$(lr 'defaultRequest.memory')" &&
    eq 200m "$(lr 'default.cpu')" && eq 128Mi "$(lr 'default.memory')"
}
applied() {
  kubectl -n $NS run probe --image=busybox:1.36 --dry-run=server -o jsonpath='{.spec.containers[0].resources.requests.cpu} {.spec.containers[0].resources.limits.memory}' --command -- sleep 1 2>/dev/null | grep -qx '100m 128Mi'
}
hard() { kubectl -n $NS get resourcequota team-quota -o jsonpath="{.spec.hard.$1}"; }
cpu_one() { [[ "$(hard 'requests\.cpu')" =~ ^(1|1000m)$ ]]; }
quota_ok() { eq 4 "$(hard pods)" && cpu_one && eq 1Gi "$(hard 'limits\.memory')"; }
rejected() { grep -q 'exceeded quota' "$COURSE/rejected.txt" && grep -q team-quota "$COURSE/rejected.txt" && ! kubectl -n $NS get pod greedy; }

check 2 "LimitRange defaults sets the container default requests and limits" defaults_ok
check 1 "A new Pod without resources gets the defaults" applied
check 2 "ResourceQuota team-quota limits pods, requests.cpu and limits.memory" quota_ok
check 2 "rejected.txt holds the quota error, and Pod greedy doesn't exist" rejected
