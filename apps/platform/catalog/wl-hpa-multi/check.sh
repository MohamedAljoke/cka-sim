hpa() { kubectl -n wl-hpa-multi get hpa report -o jsonpath="$1" 2>/dev/null; }
metric() { hpa "{.spec.metrics[?(@.resource.name==\"$1\")].resource.target}"; }
basics() { eq "Deployment/report 1-5" "$(hpa '{.spec.scaleTargetRef.kind}/{.spec.scaleTargetRef.name} {.spec.minReplicas}-{.spec.maxReplicas}')"; }
cpu_70() { metric cpu | grep -q '"averageUtilization":70' && metric cpu | grep -q '"type":"Utilization"'; }
memory_100() { metric memory | grep -q '"averageValue":"100Mi"' && metric memory | grep -q '"type":"AverageValue"'; }
reads_both() { eq 2 "$(hpa '{range .status.currentMetrics[*]}{.resource.name}{"\n"}{end}' | grep -c .)"; }
answer() { tr -d '[:space:]' < "$COURSE/decides.txt" 2>/dev/null; }

check 1 "HPA report targets Deployment report with 1 to 5 replicas" basics
check 1 "HPA report targets 70% average CPU utilization" cpu_70
check 1 "HPA report targets 100Mi average memory per Pod" memory_100
check 1 "HPA report reads both metrics" wait_for 120 reads_both
check 1 "decides.txt says highest" eq highest "$(answer)"
