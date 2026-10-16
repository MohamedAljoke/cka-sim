`spec.behavior` tunes how fast an HPA moves, separately for `scaleUp` and `scaleDown`:

```yaml
spec:
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 600
      policies:
      - type: Pods
        value: 1
        periodSeconds: 60
```

```sh
kubectl -n wl-hpa-behavior patch hpa worker --type merge -p \
  '{"spec":{"behavior":{"scaleDown":{"stabilizationWindowSeconds":600,"policies":[{"type":"Pods","value":1,"periodSeconds":60}]}}}}'
```

- **Stabilization window:** the HPA remembers its recommendations for that many seconds and,
  when scaling down, picks the highest of them. A short quiet spell then can't shrink the
  Deployment. The default for scale-down is 300 seconds; for scale-up it is 0.
- **Policies:** limit the change per period, as a number of Pods (`Pods`) or a share of the
  current replicas (`Percent`). With several policies, `selectPolicy` decides: `Max` (the
  default) allows the bigger change, `Min` the smaller one, `Disabled` none at all.

The API server fills in the scale-up defaults when you set only `scaleDown`, which is why
`kubectl get hpa worker -o yaml` shows a `scaleUp` section you didn't write.
