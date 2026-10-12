fresh_ns wl-hpa
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: wl-hpa
spec:
  replicas: 1
  selector:
    matchLabels: {app: api}
  template:
    metadata:
      labels: {app: api}
    spec:
      containers:
      - name: api
        image: nginx:1.27-alpine
        resources:
          requests: {cpu: 50m}
YAML
