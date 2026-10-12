NS=st-released
bound() { eq "Bound pv-archive" "$(kubectl -n $NS get pvc archive-v2 -o jsonpath='{.status.phase} {.spec.volumeName}')"; }
reads() { eq Running "$(kubectl -n $NS get pod reader -o jsonpath='{.status.phase}')" && eq "archived data" "$(kubectl -n $NS logs reader 2>/dev/null)"; }
kept() { bound && eq Retain "$(kubectl get pv pv-archive -o jsonpath='{.spec.persistentVolumeReclaimPolicy}')"; }
bound && wait_for 90 reads

check 3 "PVC archive-v2 is Bound to pv-archive" bound
check 2 "Pod reader runs and reads the old data" reads
check 1 "pv-archive keeps reclaim policy Retain" kept
