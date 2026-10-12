fresh_ns tr-requested-cpu
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: batch
  namespace: tr-requested-cpu
spec:
  replicas: 1
  selector:
    matchLabels:
      app: batch
  template:
    metadata:
      labels:
        app: batch
    spec:
      nodeSelector:
        kubernetes.io/hostname: cka-sim-worker
      tolerations:
      - key: dedicated
        operator: Exists
      containers:
      - name: batch
        image: busybox:1.36
        command: [sleep, 1d]
        resources:
          requests:
            cpu: 1500m
YAML
kubectl -n tr-requested-cpu rollout status deployment/batch --timeout=120s >/dev/null
