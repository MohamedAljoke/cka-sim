kubectl -n st-wffc get events --field-selector involvedObject.name=cache,reason=WaitForFirstConsumer -o jsonpath='{.items[0].message}' > /opt/course/st-wffc/reason.txt
kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: cache-user
  namespace: st-wffc
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: [sleep, 1d]
    volumeMounts:
    - name: cache
      mountPath: /cache
  volumes:
  - name: cache
    persistentVolumeClaim:
      claimName: cache
YAML
kubectl -n st-wffc wait --for=condition=Ready pod/cache-user --timeout=90s
