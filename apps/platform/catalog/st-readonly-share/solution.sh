kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: writer
  namespace: st-readonly-share
spec:
  containers:
  - name: writer
    image: busybox:1.36
    command: [sh, -c, 'while true; do date >> /logs/app.log; sleep 2; done']
    volumeMounts:
    - name: shared
      mountPath: /logs
      subPath: writer-logs
  volumes:
  - name: shared
    persistentVolumeClaim:
      claimName: shared
---
apiVersion: v1
kind: Pod
metadata:
  name: auditor
  namespace: st-readonly-share
spec:
  containers:
  - name: auditor
    image: busybox:1.36
    command: [sleep, 1d]
    volumeMounts:
    - name: shared
      mountPath: /audit
      readOnly: true
  volumes:
  - name: shared
    persistentVolumeClaim:
      claimName: shared
YAML
kubectl -n st-readonly-share wait --for=condition=Ready pod/writer pod/auditor --timeout=120s
