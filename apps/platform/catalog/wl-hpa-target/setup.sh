fresh_ns wl-hpa-target
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orders
  namespace: wl-hpa-target
spec:
  replicas: 1
  selector:
    matchLabels: {app: orders}
  template:
    metadata:
      labels: {app: orders}
    spec:
      containers:
      - name: orders
        image: nginx:1.27-alpine
        resources:
          requests: {cpu: 50m}
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: orders
  namespace: wl-hpa-target
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: ReplicaSet
    name: orders
  minReplicas: 2
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 75
YAML
