fresh_ns tr-cordon
fresh_course
kubectl cordon cka-sim-worker2 >/dev/null
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent
  namespace: tr-cordon
spec:
  replicas: 2
  selector:
    matchLabels:
      app: agent
  template:
    metadata:
      labels:
        app: agent
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                app: agent
            topologyKey: kubernetes.io/hostname
      tolerations:
      - key: dedicated
        operator: Exists
      containers:
      - name: agent
        image: busybox:1.36
        command: [sh, -c, sleep 1d]
YAML
pending() { [ -n "$(kubectl -n tr-cordon get pods -l app=agent --field-selector status.phase=Pending -o name)" ]; }
wait_for 60 pending
