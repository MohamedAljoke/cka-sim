hpa() { kubectl -n wl-hpa get hpa api -o jsonpath="$1" 2>/dev/null; }
target() { eq "Deployment/api" "$(hpa '{.spec.scaleTargetRef.kind}/{.spec.scaleTargetRef.name}')"; }
bounds() { eq "2-6" "$(hpa '{.spec.minReplicas}-{.spec.maxReplicas}')"; }
cpu_60() { eq "cpu Utilization 60" "$(hpa '{.spec.metrics[0].resource.name} {.spec.metrics[0].resource.target.type} {.spec.metrics[0].resource.target.averageUtilization}')"; }
reads_cpu() { [ -n "$(hpa '{.status.currentMetrics[0].resource.current.averageUtilization}')" ]; }
scaled() { [ "$(kubectl -n wl-hpa get deployment api -o jsonpath='{.spec.replicas}')" -ge 2 ] 2>/dev/null; }

check 1 "HPA api targets Deployment api" target
check 1 "HPA api keeps 2 to 6 replicas" bounds
check 1 "HPA api targets 60% average CPU utilization" cpu_60
check 1 "HPA api reads the Pods' CPU usage and scaled api to at least 2" wait_for 120 eval 'reads_cpu && scaled'
