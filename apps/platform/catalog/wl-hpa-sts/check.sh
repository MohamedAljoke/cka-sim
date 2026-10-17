hpa() { kubectl -n wl-hpa-sts get hpa db -o jsonpath="$1" 2>/dev/null; }
target() { eq "apps/v1 StatefulSet db" "$(hpa '{.spec.scaleTargetRef.apiVersion} {.spec.scaleTargetRef.kind} {.spec.scaleTargetRef.name}')"; }
spec() { eq "1-3 cpu 80" "$(hpa '{.spec.minReplicas}-{.spec.maxReplicas} {.spec.metrics[0].resource.name} {.spec.metrics[0].resource.target.averageUtilization}')"; }
reads_cpu() { [ -n "$(hpa '{.status.currentMetrics[0].resource.current.averageUtilization}')" ]; }
answer() { tr -d '[:space:]' < "$COURSE/$1" 2>/dev/null; }

check 1 "HPA db targets StatefulSet db" target
check 1 "HPA db keeps 1 to 3 replicas at 80% CPU" spec
check 1 "HPA db reads the CPU utilization" wait_for 120 reads_cpu
check 1 "pvcs.txt says 3" eq 3 "$(answer pvcs.txt)"
check 1 "first-removed.txt names db-2" eq db-2 "$(answer first-removed.txt)"
