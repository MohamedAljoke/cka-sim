fresh_ns nw-readiness
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: nw-readiness
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: nginx:1.27-alpine
        ports:
        - containerPort: 80
        readinessProbe:
          httpGet:
            path: /healthz
            port: 80
          periodSeconds: 3
---
apiVersion: v1
kind: Service
metadata:
  name: web
  namespace: nw-readiness
spec:
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 80
YAML
kubectl -n nw-readiness run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-readiness wait --for=condition=Ready pod/client --timeout=120s >/dev/null
running() { eq 2 "$(kubectl -n nw-readiness get pods -l app=web --field-selector status.phase=Running -o name | wc -l)"; }
wait_for 120 running
