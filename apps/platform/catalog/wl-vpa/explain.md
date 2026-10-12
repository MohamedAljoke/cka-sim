```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: api
  namespace: wl-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api
  updatePolicy:
    updateMode: "Off"
```

```bash
kubectl -n wl-vpa get vpa api                 # MODE Off, CPU and MEM once a recommendation exists
kubectl -n wl-vpa describe vpa api            # Lower Bound / Target / Upper Bound per container
```

- **`updateMode: "Off"` is recommendation-only.** The recommender writes `status.recommendation`,
  and nothing touches the Pods. Quote `"Off"`, or YAML reads it as the boolean `false`.
- **Other modes act on the Pods:** `Initial` sets requests only when a Pod is created; `Recreate` (and
  `Auto`) evicts Pods to apply new requests; `InPlaceOrRecreate` resizes running Pods where it can.
  They need the VPA updater and admission controller. This lab installs only the recommender.
- **`target` is what to set as requests.** `lowerBound` and `upperBound` are the range the updater
  tolerates before acting. An idle app sits at the recommender's floor (25m CPU, 250Mi memory).
- **Don't point a VPA and an HPA at the same resource metric** (CPU or memory): they'd fight over the
  same Pods.

The VPA is a separate project (kubernetes/autoscaler), so it only exists where it's been installed.
Check first: `kubectl api-resources | grep -i verticalpodautoscaler`.
