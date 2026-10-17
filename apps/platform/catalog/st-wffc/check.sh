NS=st-wffc
reason_ok() { grep -qi 'first consumer' "$COURSE/reason.txt" 2>/dev/null; }
bound() { eq Bound "$(kubectl -n $NS get pvc cache -o jsonpath='{.status.phase}')"; }
runs() { bound && eq Running "$(kubectl -n $NS get pod cache-user -o jsonpath='{.status.phase}' 2>/dev/null)" && eq cache "$(kubectl -n $NS get pod cache-user -o jsonpath='{.spec.volumes[?(@.persistentVolumeClaim)].persistentVolumeClaim.claimName}')"; }
kubectl -n $NS get pod cache-user >/dev/null 2>&1 && wait_for 90 runs

check 2 "reason.txt holds the WaitForFirstConsumer message" reason_ok
check 2 "PVC cache is Bound" bound
check 1 "Pod cache-user runs with PVC cache" runs
