NS=wl-pod-failure-policy
job() { kubectl -n $NS get job importer -o jsonpath="$1" 2>/dev/null; }
rule() { job '{.spec.podFailurePolicy.rules[?(@.action=="FailJob")].onExitCodes}'; }
policy_ok() { eq 6 "$(job '{.spec.backoffLimit}')" && rule | jq -e '.values == [42] and .operator == "In" and ((.containerName // "importer") == "importer")' >/dev/null; }
failed_fast() { policy_ok && eq PodFailurePolicy "$(job '{.status.conditions[?(@.type=="Failed")].reason}')"; }
one_pod() { failed_fast && eq 1 "$(kubectl -n $NS get pods -l job-name=importer -o name | wc -l)"; }
policy_ok && wait_for 60 failed_fast

check 2 "importer has backoffLimit 6 and a FailJob rule for exit code 42" policy_ok
check 2 "importer failed with reason PodFailurePolicy" failed_fast
check 1 "It never retried: exactly 1 Pod ran" one_pod
