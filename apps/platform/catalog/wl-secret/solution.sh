kubectl -n wl-secret create secret generic db-creds --from-literal=username=admin --from-literal='password=S3cr3t!'
kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: db-client
  namespace: wl-secret
spec:
  containers:
  - name: db-client
    image: busybox:1.36
    command: [sleep, 1d]
    envFrom:
    - secretRef:
        name: db-creds
    volumeMounts:
    - name: creds
      mountPath: /etc/db-creds
      readOnly: true
  volumes:
  - name: creds
    secret:
      secretName: db-creds
YAML
kubectl -n wl-secret wait --for=condition=Ready pod/db-client --timeout=60s
