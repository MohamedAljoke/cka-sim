hpa() { kubectl -n wl-hpa-behavior get hpa worker -o jsonpath="$1" 2>/dev/null; }
window() { eq 600 "$(hpa '{.spec.behavior.scaleDown.stabilizationWindowSeconds}')"; }
one_per_minute() { hpa '{range .spec.behavior.scaleDown.policies[*]}{.type} {.value} {.periodSeconds}{"\n"}{end}' | grep -qx 'Pods 1 60'; }
only_that_policy() { eq 1 "$(hpa '{range .spec.behavior.scaleDown.policies[*]}{.type}{"\n"}{end}' | grep -c .)"; }
unchanged() { eq "1-8 cpu 70 Deployment/worker" "$(hpa '{.spec.minReplicas}-{.spec.maxReplicas} {.spec.metrics[0].resource.name} {.spec.metrics[0].resource.target.averageUtilization} {.spec.scaleTargetRef.kind}/{.spec.scaleTargetRef.name}')"; }
scale_up_default() { eq 0 "$(hpa '{.spec.behavior.scaleUp.stabilizationWindowSeconds}')"; }

check 2 "scale-down stabilization window is 600 seconds" window
check 2 "scale-down removes at most 1 Pod per 60 seconds" eval 'one_per_minute && only_that_policy'
check 1 "bounds, metric, target and scale-up are unchanged" eval 'unchanged && scale_up_default'
