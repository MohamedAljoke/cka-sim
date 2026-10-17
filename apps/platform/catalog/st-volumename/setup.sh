fresh_ns st-volumename
for pv in pv-orders-a pv-orders-b; do
  kubectl delete pv "$pv" --ignore-not-found --wait=true >/dev/null
  kubectl apply -f - >/dev/null <<YAML
apiVersion: v1
kind: PersistentVolume
metadata:
  name: $pv
spec:
  capacity:
    storage: 1Gi
  accessModes: [ReadWriteOnce]
  persistentVolumeReclaimPolicy: Retain
  storageClassName: orders
  hostPath:
    path: /tmp/$pv
    type: DirectoryOrCreate
YAML
done
