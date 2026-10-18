kubectl apply -f - <<'YAML'
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: log-agent
  namespace: tr-log-agent
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
      - key: dedicated
        operator: Exists
      containers:
      - name: agent
        image: busybox:1.36
        command: [sh, -c, 'tail -n+1 -F /var/log/pods/tr-log-agent-apps_*/*/*.log']
        volumeMounts:
        - name: pods
          mountPath: /var/log/pods
          readOnly: true
      volumes:
      - name: pods
        hostPath:
          path: /var/log/pods
YAML
kubectl -n tr-log-agent rollout status daemonset log-agent --timeout=120s
