fresh_ns nw-clusterip
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: nw-clusterip
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
        command: [sh, -c, 'mkdir -p /www && echo "api ok" > /www/index.html && httpd -f -p 8080 -h /www']
        ports:
        - containerPort: 8080
YAML
kubectl -n nw-clusterip run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-clusterip rollout status deployment/api --timeout=120s >/dev/null
kubectl -n nw-clusterip wait --for=condition=Ready pod/client --timeout=120s >/dev/null
