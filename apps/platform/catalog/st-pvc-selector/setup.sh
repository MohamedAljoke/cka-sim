fresh_ns st-pvc-selector
fresh_course
for pv in ssd:2Gi hdd:5Gi; do
  kubectl delete pv "pv-${pv%%:*}" --ignore-not-found --wait=true >/dev/null
  kubectl apply -f - >/dev/null <<YAML
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-${pv%%:*}
  labels:
    disk: ${pv%%:*}
spec:
  capacity:
    storage: ${pv#*:}
  accessModes: [ReadWriteOnce]
  persistentVolumeReclaimPolicy: Retain
  storageClassName: tiered
  hostPath:
    path: /tmp/st-pvc-selector-${pv%%:*}
    type: DirectoryOrCreate
YAML
done
