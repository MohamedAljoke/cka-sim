fresh_ns tr-crashloop
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-config
  namespace: tr-crashloop
data:
  config.yml: |
    port: 8080
    log_level: info
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: tr-crashloop
spec:
  replicas: 1
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
      - name: api
        image: busybox:1.36
        command:
        - sh
        - -c
        - |
          echo "api starting"
          echo "reading /etc/api/config.yaml"
          [ -f /etc/api/config.yaml ] || { echo "FATAL: /etc/api/config.yaml not found"; exit 1; }
          echo "config loaded, serving"
          touch /tmp/ready
          sleep 1d
        readinessProbe:
          exec:
            command: [cat, /tmp/ready]
          periodSeconds: 2
        volumeMounts:
        - name: config
          mountPath: /etc/api
      volumes:
      - name: config
        configMap:
          name: api-config
YAML
backing_off() { eq CrashLoopBackOff "$(kubectl -n tr-crashloop get pods -l app=api -o jsonpath='{.items[0].status.containerStatuses[0].state.waiting.reason}')"; }
wait_for 90 backing_off
