fresh_ns wl-canary
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-stable
  namespace: wl-canary
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web
      track: stable
  template:
    metadata:
      labels:
        app: web
        track: stable
    spec:
      containers:
      - name: web
        image: busybox:1.36
        command: [sh, -c, 'mkdir -p /www && echo "$VERSION" > /www/index.html && httpd -f -p 8080 -h /www']
        env:
        - name: VERSION
          value: stable
---
apiVersion: v1
kind: Service
metadata:
  name: web
  namespace: wl-canary
spec:
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 8080
YAML
kubectl -n wl-canary run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n wl-canary rollout status deployment/web-stable --timeout=120s >/dev/null
kubectl -n wl-canary wait --for=condition=Ready pod/client --timeout=120s >/dev/null
