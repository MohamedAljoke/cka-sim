fresh_ns ar-finalizer
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data
  namespace: ar-finalizer
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: standard
  resources:
    requests:
      storage: 50Mi
---
apiVersion: v1
kind: Pod
metadata:
  name: legacy-app
  namespace: ar-finalizer
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: [sleep, 1d]
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: data
YAML
kubectl -n ar-finalizer wait --for=condition=Ready pod/legacy-app --timeout=120s >/dev/null
kubectl -n ar-finalizer delete pvc data --wait=false >/dev/null
