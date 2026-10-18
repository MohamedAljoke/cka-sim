kubectl -n wl-configmap create configmap app-config --from-literal=LOG_LEVEL=debug --from-file=/opt/course/wl-configmap/app.properties
kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: app
  namespace: wl-configmap
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: [sleep, 1d]
    env:
    - name: LOG_LEVEL
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: LOG_LEVEL
    volumeMounts:
    - name: config
      mountPath: /etc/app
  volumes:
  - name: config
    configMap:
      name: app-config
      items:
      - key: app.properties
        path: app.properties
YAML
kubectl -n wl-configmap wait --for=condition=Ready pod/app --timeout=60s
