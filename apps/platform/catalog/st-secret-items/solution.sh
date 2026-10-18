kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: web
  namespace: st-secret-items
spec:
  containers:
  - name: web
    image: busybox:1.36
    command: [sleep, 1d]
    volumeMounts:
    - name: key
      mountPath: /etc/ssl/private
      readOnly: true
  volumes:
  - name: key
    secret:
      secretName: tls-bundle
      items:
      - key: tls.key
        path: server.key
        mode: 0400
YAML
kubectl -n st-secret-items wait --for=condition=Ready pod/web --timeout=60s
