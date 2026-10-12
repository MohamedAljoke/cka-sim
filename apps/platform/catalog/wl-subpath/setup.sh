fresh_ns wl-subpath
kubectl -n wl-subpath create configmap app-config --from-literal=settings.conf=level=info >/dev/null
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: wl-subpath
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: [sleep, 1d]
        volumeMounts:
        - name: config
          mountPath: /etc/app/settings.conf
          subPath: settings.conf
      volumes:
      - name: config
        configMap:
          name: app-config
YAML
kubectl -n wl-subpath rollout status deployment/app --timeout=120s >/dev/null
kubectl -n wl-subpath create configmap app-config --from-literal=settings.conf=level=debug --dry-run=client -o yaml | kubectl replace -f - >/dev/null
