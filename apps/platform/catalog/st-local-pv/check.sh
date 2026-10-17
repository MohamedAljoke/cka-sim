NS=st-local-pv
bound() { eq "Bound local-data" "$(kubectl -n $NS get pvc data -o jsonpath='{.status.phase} {.spec.volumeName}')"; }
runs() { eq "Running cka-sim-worker2" "$(kubectl -n $NS get pod app -o jsonpath='{.status.phase} {.spec.nodeName}' 2>/dev/null)"; }
wrote() { runs && [ -s /mnt/local-data/starts ]; }
pv_kept() { bound && eq "Retain local-storage /mnt/local-data" "$(kubectl get pv local-data -o jsonpath='{.spec.persistentVolumeReclaimPolicy} {.spec.storageClassName} {.spec.local.path}')"; }
kubectl -n $NS get pod app >/dev/null 2>&1 && [ "$(kubectl -n $NS get pod app -o jsonpath="{.spec.nodeSelector.kubernetes\.io/hostname}")" != cka-sim-worker ] && wait_for 60 runs

check 2 "PVC data is Bound to local-data" bound
check 2 "Pod app runs on cka-sim-worker2" runs
check 1 "It writes to the local disk at /mnt/local-data" wrote
check 1 "PV local-data was not modified" pv_kept
