pv() { k get pv data-pv -o jsonpath="$1"; }
pod_ready() { k -n storage wait pod/writer --for=condition=Ready --timeout=5s; }
wait_for 90 pod_ready
check 2 "data-pv: 1Gi, RWO, Retain, class manual, hostPath /mnt/data" \
  eq "1Gi ReadWriteOnce Retain manual /mnt/data" "$(pv '{.spec.capacity.storage} {.spec.accessModes[0]} {.spec.persistentVolumeReclaimPolicy} {.spec.storageClassName} {.spec.hostPath.path}')"
check 2 "data-pvc is Bound to data-pv" eq "Bound data-pv" "$(k -n storage get pvc data-pvc -o jsonpath='{.status.phase} {.spec.volumeName}')"
check 1 "Pod writer is Running" pod_ready
check 2 "writer wrote /data/ok through the claim" eq ok "$(k -n storage exec writer -- cat /data/ok)"
