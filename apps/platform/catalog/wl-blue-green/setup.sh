fresh_ns wl-blue-green
deploy() {
  kubectl apply -f - >/dev/null <<YAML
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-$1
  namespace: wl-blue-green
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
      version: $1
  template:
    metadata:
      labels:
        app: web
        version: $1
    spec:
      containers:
      - name: web
        image: busybox:1.36
        command: [sh, -c, 'mkdir -p /www && echo $1 > /www/index.html && httpd -f -p 8080 -h /www']
YAML
}
deploy blue
deploy green
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Service
metadata:
  name: web
  namespace: wl-blue-green
spec:
  selector:
    app: web
    version: blue
  ports:
  - port: 80
    targetPort: 8080
YAML
kubectl -n wl-blue-green run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n wl-blue-green rollout status deployment/web-blue --timeout=120s >/dev/null
kubectl -n wl-blue-green rollout status deployment/web-green --timeout=120s >/dev/null
kubectl -n wl-blue-green wait --for=condition=Ready pod/client --timeout=120s >/dev/null
