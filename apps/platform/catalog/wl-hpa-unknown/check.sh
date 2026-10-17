hpa() { kubectl -n wl-hpa-unknown get hpa shop -o jsonpath="$1" 2>/dev/null; }
requests() { kubectl -n wl-hpa-unknown get deployment shop -o jsonpath='{range .spec.template.spec.containers[*]}{.name}={.resources.requests.cpu}{"\n"}{end}' | sort | xargs; }
unchanged() { eq "1-4 cpu 50 Deployment/shop" "$(hpa '{.spec.minReplicas}-{.spec.maxReplicas} {.spec.metrics[0].resource.name} {.spec.metrics[0].resource.target.averageUtilization} {.spec.scaleTargetRef.kind}/{.spec.scaleTargetRef.name}')"; }
reads_cpu() { [ -n "$(hpa '{.status.currentMetrics[0].resource.current.averageUtilization}')" ]; }

check 2 "both containers of shop request 50m CPU" eq "metrics=50m web=50m" "$(requests)"
check 3 "HPA shop is unchanged and shows the real CPU utilization" eval 'unchanged && wait_for 120 reads_cpu'
