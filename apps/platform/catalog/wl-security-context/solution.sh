kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: secure
  namespace: wl-security-context
spec:
  securityContext:
    runAsUser: 1000
    runAsGroup: 3000
  containers:
  - name: secure
    image: busybox:1.36
    command: [sleep, 1d]
    securityContext:
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
      capabilities:
        drop: [ALL]
    volumeMounts:
    - name: tmp
      mountPath: /tmp
  volumes:
  - name: tmp
    emptyDir: {}
YAML
kubectl -n wl-security-context wait --for=condition=Ready pod/secure --timeout=60s
