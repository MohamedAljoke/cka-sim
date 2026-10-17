fresh_ns st-pending-pvc
kubectl delete pv pv-reports --ignore-not-found --wait=true >/dev/null
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-reports
spec:
  capacity:
    storage: 1Gi
  accessModes: [ReadWriteOnce]
  persistentVolumeReclaimPolicy: Retain
  storageClassName: manual
  hostPath:
    path: /tmp/st-pending-pvc
    type: DirectoryOrCreate
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: reports
  namespace: st-pending-pvc
spec:
  accessModes: [ReadWriteMany]
  storageClassName: manual
  resources:
    requests:
      storage: 2Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: reporter
  namespace: st-pending-pvc
spec:
  containers:
  - name: reporter
    image: busybox:1.36
    command: [sh, -c, 'date >> /reports/runs.log; sleep 1d']
    volumeMounts:
    - name: reports
      mountPath: /reports
  volumes:
  - name: reports
    persistentVolumeClaim:
      claimName: reports
YAML
