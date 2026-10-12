fresh_ns tr-oomkilled
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cache
  namespace: tr-oomkilled
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cache
  template:
    metadata:
      labels:
        app: cache
    spec:
      containers:
      - name: cache
        image: busybox:1.36
        command: [sh, -c, 'echo loading data; head -c 50m /dev/zero | tail > /dev/null; echo ready; sleep 1d']
        resources:
          requests:
            memory: 16Mi
          limits:
            memory: 32Mi
YAML
oom() { kubectl -n tr-oomkilled get pods -l app=cache -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated.reason}' | grep -q OOMKilled; }
wait_for 120 oom
