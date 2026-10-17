fresh_ns wl-resize
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: worker
  namespace: wl-resize
spec:
  containers:
  - name: worker
    image: busybox:1.36
    command: [sleep, 1d]
    resources:
      requests:
        cpu: 100m
        memory: 32Mi
      limits:
        cpu: 200m
        memory: 64Mi
YAML
kubectl -n wl-resize wait --for=condition=Ready pod/worker --timeout=120s >/dev/null
