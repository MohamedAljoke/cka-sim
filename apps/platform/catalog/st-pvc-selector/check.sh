pvc() { kubectl -n st-pvc-selector get pvc fast -o jsonpath="$1" 2>/dev/null; }
selector_ok() { eq ssd "$(pvc '{.spec.selector.matchLabels.disk}')" || pvc '{.spec.selector.matchExpressions}' | grep -q ssd; }
bound() { selector_ok && eq "Bound pv-ssd" "$(pvc '{.status.phase} {.spec.volumeName}')"; }
capacity_ok() { bound && eq 2Gi "$(tr -d ' \r\n' < "$COURSE/capacity.txt" 2>/dev/null)"; }
selector_ok && wait_for 30 bound

check 2 "PVC fast selects disk=ssd and is Bound to pv-ssd" bound
check 2 "capacity.txt says 2Gi" capacity_ok
