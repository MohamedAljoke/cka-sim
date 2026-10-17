pod() { kubectl -n wl-hpa-load get pod load -o jsonpath="$1" 2>/dev/null; }
load_runs() { eq "busybox:1.36 Running" "$(pod '{.spec.containers[0].image} {.status.phase}')"; }
scaled() { [ "$(kubectl -n wl-hpa-load get deployment web -o jsonpath='{.spec.replicas}')" -ge 2 ] 2>/dev/null; }
answer() { tr -d '[:space:]' < "$COURSE/scale-down.txt" 2>/dev/null; }

check 1 "Pod load runs busybox:1.36" load_runs
check 3 "HPA web scaled Deployment web to at least 2 replicas" wait_for 120 scaled
check 2 "scale-down.txt says 420" eq 420 "$(answer)"
