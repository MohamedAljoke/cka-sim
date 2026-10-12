fresh_ns st-local-pv
fresh_course
kubectl delete pv local-data --ignore-not-found --wait=true >/dev/null
mkdir -p /mnt/local-data
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-data
spec:
  capacity:
    storage: 1Gi
  accessModes: [ReadWriteOnce]
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /mnt/local-data
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values: [cka-sim-worker2]
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data
  namespace: st-local-pv
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: local-storage
  resources:
    requests:
      storage: 1Gi
YAML
cat > "$COURSE/app.yaml" <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: app
  namespace: st-local-pv
spec:
  nodeSelector:
    kubernetes.io/hostname: cka-sim-worker
  tolerations:
  - key: dedicated
    operator: Exists
  containers:
  - name: app
    image: busybox:1.36
    command: [sh, -c, 'date >> /data/starts; sleep 1d']
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: data
YAML
chmod a+rw "$COURSE/app.yaml"
kubectl apply -f "$COURSE/app.yaml" >/dev/null
unschedulable() { kubectl -n st-local-pv get events --field-selector involvedObject.name=app,reason=FailedScheduling -o name | grep -q .; }
wait_for 60 unschedulable
