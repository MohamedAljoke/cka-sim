fresh_ns wl-hpa-sts
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Service
metadata:
  name: db
  namespace: wl-hpa-sts
spec:
  clusterIP: None
  selector: {app: db}
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: db
  namespace: wl-hpa-sts
spec:
  serviceName: db
  replicas: 3
  selector:
    matchLabels: {app: db}
  template:
    metadata:
      labels: {app: db}
    spec:
      containers:
      - name: db
        image: busybox:1.36
        command: [sleep, 1d]
        resources:
          requests: {cpu: 20m}
        volumeMounts:
        - name: data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ReadWriteOnce]
      resources:
        requests: {storage: 64Mi}
YAML
ready() { eq 3 "$(kubectl -n wl-hpa-sts get statefulset db -o jsonpath='{.status.readyReplicas}')"; }
wait_for 150 ready
