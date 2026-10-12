fresh_ns wl-quota-rollout
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-cpu
  namespace: wl-quota-rollout
spec:
  hard:
    requests.cpu: "1"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: wl-quota-rollout
spec:
  replicas: 2
  progressDeadlineSeconds: 20
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
      - name: nginx
        image: nginx:1.26-alpine
        resources:
          requests:
            cpu: 400m
YAML
kubectl -n wl-quota-rollout rollout status deployment/api --timeout=120s >/dev/null
kubectl -n wl-quota-rollout set image deployment/api nginx=nginx:1.27-alpine >/dev/null
stuck() { eq ProgressDeadlineExceeded "$(kubectl -n wl-quota-rollout get deployment api -o jsonpath='{.status.conditions[?(@.type=="Progressing")].reason}')"; }
wait_for 90 stuck
