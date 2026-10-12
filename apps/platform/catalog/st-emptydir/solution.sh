pod() {
  kubectl apply -f - <<YAML
apiVersion: v1
kind: Pod
metadata:
  name: $1
  namespace: st-emptydir
spec:
  containers:
  - name: writer
    image: busybox:1.36
    command: [sh, -c, 'while true; do date >> /data/out.txt; sleep 2; done']
    volumeMounts:
    - {name: data, mountPath: /data}
  - name: reader
    image: busybox:1.36
    command: [sh, -c, 'sleep 1d']
    volumeMounts:
    - {name: data, mountPath: /data}
  volumes:
  - name: data
    emptyDir: $2
YAML
}
pod share '{}'
pod share-mem '{medium: Memory, sizeLimit: 64Mi}'
kubectl -n st-emptydir wait --for=condition=Ready pod/share pod/share-mem --timeout=90s
