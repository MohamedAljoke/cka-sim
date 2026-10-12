`kubectl autoscale` can set both targets (`--cpu=70% --memory=100Mi`), but writing the YAML makes
the two target types plain:

```yaml
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
```

- `Utilization` is a percentage of the Pods' requests, so it needs a request for that resource.
- `AverageValue` is an absolute amount per Pod, so it doesn't.

With several metrics the HPA computes a replica count for each one and uses the **highest**. No
metric is ever starved: if memory needs 4 Pods and CPU needs 2, you get 4. Scale-down only
happens when every metric agrees that fewer Pods are enough.

Memory is a weak autoscaling signal for many apps: memory often isn't released when load drops,
so the HPA may never scale back down.
