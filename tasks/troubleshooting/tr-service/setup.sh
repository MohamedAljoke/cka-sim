fresh_ns shop
k apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: shop
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
      tier: frontend
  template:
    metadata:
      labels:
        app: web
        tier: frontend
    spec:
      containers:
      - name: nginx
        image: nginx:1.27-alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: web
  namespace: shop
spec:
  selector:
    app: webapp
  ports:
  - port: 80
    targetPort: 8080
YAML
k -n shop rollout status deploy/web --timeout=180s >/dev/null
remember generation "$(k -n shop get deploy web -o jsonpath='{.metadata.generation}')"
