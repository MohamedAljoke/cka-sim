fresh_ns nw-named-port
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: nw-named-port
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
        image: busybox:1.36
        command: [sh, -c, 'mkdir -p /www && echo "api on $PORT" > /www/index.html && httpd -f -p "$PORT" -h /www']
        env:
        - name: PORT
          value: "8080"
        ports:
        - name: http
          containerPort: 8080
YAML
kubectl -n nw-named-port run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-named-port rollout status deployment/api --timeout=120s >/dev/null
kubectl -n nw-named-port wait --for=condition=Ready pod/client --timeout=120s >/dev/null
