fresh_ns st-emptydir-to-pvc
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: notes
  namespace: st-emptydir-to-pvc
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: notes
  template:
    metadata:
      labels:
        app: notes
    spec:
      containers:
      - name: notes
        image: busybox:1.36
        command: [sh, -c, 'echo "started by $(hostname)" >> /data/notes.txt; sleep 1d']
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        emptyDir: {}
YAML
kubectl -n st-emptydir-to-pvc rollout status deployment/notes --timeout=120s >/dev/null
