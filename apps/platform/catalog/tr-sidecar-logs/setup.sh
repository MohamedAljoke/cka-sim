fresh_ns tr-sidecar-logs
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orders
  namespace: tr-sidecar-logs
spec:
  replicas: 1
  selector:
    matchLabels:
      app: orders
  template:
    metadata:
      labels:
        app: orders
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: [sh, -c, 'i=0; while true; i=$((i+1)); do echo "$(date +%T) processed order $i" >> /var/log/orders/orders.log; sleep 2; done']
        volumeMounts:
        - name: logs
          mountPath: /var/log/orders
      volumes:
      - name: logs
        emptyDir: {}
YAML
kubectl -n tr-sidecar-logs rollout status deployment/orders --timeout=120s >/dev/null
