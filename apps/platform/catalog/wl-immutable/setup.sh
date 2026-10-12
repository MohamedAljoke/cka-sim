fresh_ns wl-immutable
kubectl -n wl-immutable create configmap app-config-v1 --from-literal=color=green >/dev/null
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: wl-immutable
spec:
  replicas: 2
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
        command: [sh, -c, 'echo "COLOR=$COLOR"; sleep 1d']
        env:
        - name: COLOR
          valueFrom:
            configMapKeyRef:
              name: app-config-v1
              key: color
YAML
kubectl -n wl-immutable rollout status deployment/app --timeout=120s >/dev/null
