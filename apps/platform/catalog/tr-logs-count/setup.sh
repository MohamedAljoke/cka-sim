fresh_ns tr-logs-count
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: worker
  namespace: tr-logs-count
spec:
  replicas: 3
  selector:
    matchLabels:
      app: worker
  template:
    metadata:
      labels:
        app: worker
    spec:
      containers:
      - name: worker
        image: busybox:1.36
        command:
        - sh
        - -c
        - |
          n=$((RANDOM % 6))
          i=0
          while [ $i -lt 8 ]; do
            echo "INFO job $i done"
            [ $i -lt $n ] && echo "ERROR job $i failed: timeout"
            i=$((i+1))
          done
          sleep 1d
YAML
kubectl -n tr-logs-count rollout status deployment/worker --timeout=120s >/dev/null
