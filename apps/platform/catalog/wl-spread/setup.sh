fresh_ns wl-spread
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: wl-spread
spec:
  replicas: 4
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            preference:
              matchExpressions:
              - key: kubernetes.io/hostname
                operator: In
                values: [cka-sim-worker2]
      tolerations:
      - key: dedicated
        operator: Exists
      containers:
      - name: nginx
        image: nginx:1.27-alpine
YAML
kubectl -n wl-spread rollout status deployment/web --timeout=120s >/dev/null
