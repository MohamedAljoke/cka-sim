kubectl apply -f - <<'YAML'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: orders
  namespace: st-volumename
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: orders
  volumeName: pv-orders-b
  resources:
    requests:
      storage: 1Gi
YAML
