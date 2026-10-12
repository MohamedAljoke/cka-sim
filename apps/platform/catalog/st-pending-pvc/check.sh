NS=st-pending-pvc
pvc() { kubectl -n $NS get pvc reports -o jsonpath="$1" 2>/dev/null; }
bound() { eq "Bound pv-reports" "$(pvc '{.status.phase} {.spec.volumeName}')"; }
runs() { eq Running "$(kubectl -n $NS get pod reporter -o jsonpath='{.status.phase}' 2>/dev/null)"; }
pv_kept() { bound && eq '1Gi ["ReadWriteOnce"] manual Retain' "$(kubectl get pv pv-reports -o jsonpath='{.spec.capacity.storage} {.spec.accessModes} {.spec.storageClassName} {.spec.persistentVolumeReclaimPolicy}')"; }
bound && wait_for 90 runs

check 3 "PVC reports is Bound to pv-reports" bound
check 2 "Pod reporter is Running" runs
check 1 "PV pv-reports was not modified" pv_kept
