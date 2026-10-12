kubectl apply -f - >/dev/null <<'YAML'
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: report
  namespace: wl-hpa-multi
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: report
  minReplicas: 1
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: AverageValue
        averageValue: 100Mi
YAML
echo highest > /opt/course/wl-hpa-multi/decides.txt
