kubectl apply -f - <<'YAML'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data
  namespace: st-dynamic
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: standard
  resources:
    requests:
      storage: 200Mi
---
apiVersion: v1
kind: Pod
metadata:
  name: writer
  namespace: st-dynamic
spec:
  containers:
  - name: writer
    image: busybox:1.36
    command: [sh, -c, 'echo hello > /data/hello; sleep 1d']
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: data
YAML
kubectl -n st-dynamic wait --for=condition=Ready pod/writer --timeout=120s
pv=$(kubectl -n st-dynamic get pvc data -o jsonpath='{.spec.volumeName}')
echo "$pv" > /opt/course/st-dynamic/pv.txt
kubectl get pv "$pv" -o jsonpath='{.spec.persistentVolumeReclaimPolicy}' > /opt/course/st-dynamic/policy.txt
kubectl -n st-dynamic delete pod writer --grace-period=1
kubectl -n st-dynamic delete pvc data
kubectl wait --for=delete pv/"$pv" --timeout=60s
