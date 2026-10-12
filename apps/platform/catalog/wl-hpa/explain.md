`kubectl autoscale` writes an `autoscaling/v2` HPA:

```sh
kubectl -n wl-hpa autoscale deployment api --min=2 --max=6 --cpu=60%
```

The same object as YAML:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api
  namespace: wl-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api
  minReplicas: 2
  maxReplicas: 6
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 60
```

Utilization is a percentage of the containers' CPU **requests**, so it only works because `api`
requests `50m`. The HPA raises the Deployment to `minReplicas` at once, then every 15 seconds
compares the measured average with the target:
`desired = ceil(current replicas × current / target)`.

Older kubectl versions used `--cpu-percent=60`; it still works but is deprecated.
