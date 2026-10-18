kubectl apply -f - <<'YAML'
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: log-agent
  namespace: wl-daemonset
spec:
  selector:
    matchLabels:
      app: log-agent
  template:
    metadata:
      labels:
        app: log-agent
    spec:
      tolerations:
      - operator: Exists
      containers:
      - name: log-agent
        image: busybox:1.36
        command: [sleep, 1d]
YAML
kubectl -n wl-daemonset rollout status daemonset log-agent --timeout=120s
