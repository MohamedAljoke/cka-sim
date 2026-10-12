fresh_ns wl-hpa-multi
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: report
  namespace: wl-hpa-multi
spec:
  replicas: 1
  selector:
    matchLabels: {app: report}
  template:
    metadata:
      labels: {app: report}
    spec:
      containers:
      - name: report
        image: nginx:1.27-alpine
        resources:
          requests: {cpu: 50m, memory: 32Mi}
YAML
