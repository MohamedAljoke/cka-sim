hpa() { kubectl -n wl-hpa-target get hpa orders -o jsonpath="$1" 2>/dev/null; }
target() { eq "apps/v1 Deployment orders" "$(hpa '{.spec.scaleTargetRef.apiVersion} {.spec.scaleTargetRef.kind} {.spec.scaleTargetRef.name}')"; }
kept() { eq "2-5 75" "$(hpa '{.spec.minReplicas}-{.spec.maxReplicas} {.spec.metrics[0].resource.target.averageUtilization}')"; }
able() { eq True "$(hpa '{.status.conditions[?(@.type=="AbleToScale")].status}')"; }
reads_cpu() { [ -n "$(hpa '{.status.currentMetrics[0].resource.current.averageUtilization}')" ]; }
scaled() { [ "$(kubectl -n wl-hpa-target get deployment orders -o jsonpath='{.spec.replicas}')" -ge 2 ] 2>/dev/null; }

check 1 "HPA orders targets Deployment orders" target
check 3 "HPA orders keeps 2 to 5 replicas at 75% CPU, can scale and shows the CPU utilization" eval 'kept && wait_for 120 eval "able && reads_cpu && scaled"'
