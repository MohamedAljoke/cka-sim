fresh_ns st-node-paths
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: cache
  namespace: st-node-paths
spec:
  nodeName: cka-sim-control-plane
  containers:
  - name: cache
    image: busybox:1.36
    command: [sh, -c, 'echo warm > /scratch/state; sleep 1d']
    volumeMounts:
    - name: scratch
      mountPath: /scratch
  volumes:
  - name: scratch
    emptyDir: {}
YAML
kubectl -n st-node-paths wait --for=condition=Ready pod/cache --timeout=120s >/dev/null
