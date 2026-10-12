fresh_ns nw-netpol-ports
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: server
  namespace: nw-netpol-ports
  labels:
    app: server
spec:
  containers:
  - name: server
    image: busybox:1.36
    command: [sh, -c, 'for p in 8000 8005 8010 8011 9100; do httpd -p $p; done; sleep 1d']
    ports:
    - name: metrics
      containerPort: 9100
YAML
kubectl -n nw-netpol-ports run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-ports wait --for=condition=Ready pod --all --timeout=120s >/dev/null
