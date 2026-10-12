replicas() { kubectl -n wl-hpa-fight get deployment web -o jsonpath='{.spec.replicas}'; }
hpa_bounds() { eq "3-6" "$(kubectl -n wl-hpa-fight get hpa web -o jsonpath='{.spec.minReplicas}-{.spec.maxReplicas}')"; }
at_least() { [ "$(replicas)" -ge "$1" ] 2>/dev/null; }
no_replicas_field() { ! grep -qE '^\s*replicas:' "$COURSE/web.yaml"; }
survives_apply() {
  hpa_bounds || return 1
  wait_for 60 at_least 3 || return 1
  local before
  before=$(replicas)
  kubectl apply -f "$COURSE/web.yaml" >/dev/null || return 1
  eq "$before" "$(replicas)"
}

check 2 "web.yaml no longer sets the replica count" no_replicas_field
check 4 "kubectl apply -f web.yaml keeps the replica count of HPA web (3 to 6)" survives_apply
