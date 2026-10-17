fresh_ns st-released
kubectl delete pv pv-archive --ignore-not-found --wait=true >/dev/null
claim() {
  kubectl apply -f - >/dev/null <<YAML
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: $1
  namespace: st-released
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: manual
  resources:
    requests:
      storage: 1Gi
YAML
}
pod() {
  kubectl apply -f - >/dev/null <<YAML
apiVersion: v1
kind: Pod
metadata:
  name: $1
  namespace: st-released
spec:
  containers:
  - name: main
    image: busybox:1.36
    command: [sh, -c, '$3; sleep 1d']
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: $2
YAML
}
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-archive
spec:
  capacity:
    storage: 1Gi
  accessModes: [ReadWriteOnce]
  persistentVolumeReclaimPolicy: Retain
  storageClassName: manual
  hostPath:
    path: /tmp/st-released
    type: DirectoryOrCreate
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values: [cka-sim-worker2]
YAML
claim archive
pod writer archive 'echo "archived data" > /data/keep.txt'
kubectl -n st-released wait --for=condition=Ready pod/writer --timeout=120s >/dev/null
kubectl -n st-released delete pod writer --grace-period=1 >/dev/null
kubectl -n st-released delete pvc archive >/dev/null
released() { eq Released "$(kubectl get pv pv-archive -o jsonpath='{.status.phase}')"; }
wait_for 60 released
claim archive-v2
pod reader archive-v2 'cat /data/keep.txt'
