fresh_ns wl-vpa
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: wl-vpa
spec:
  replicas: 2
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
      - name: api
        image: nginx:1.27-alpine
        resources:
          requests:
            cpu: 100m
            memory: 64Mi
YAML
kubectl -n wl-vpa rollout status deployment api --timeout=120s >/dev/null
