A PVC's `selector` narrows the PVs it may bind to by their labels, on top of the usual class, mode
and size rules:

```yaml
spec:
  storageClassName: tiered
  selector:
    matchLabels:
      disk: ssd
  resources:
    requests:
      storage: 1Gi
```

Binding is all-or-nothing: a claim gets a **whole** PV. Asking for `1Gi` and binding to a 2Gi PV
gives the claim 2Gi. `spec.resources.requests` stays `1Gi`, while `status.capacity` shows `2Gi`:

```sh
kubectl -n st-pvc-selector get pvc fast            # CAPACITY 2Gi
kubectl -n st-pvc-selector get pvc fast -o jsonpath='{.spec.resources.requests.storage} {.status.capacity.storage}'
```

Selectors only work with statically created PVs. Dynamic provisioning can't honour them, so a
claim with a selector on a provisioner-backed class just stays `Pending`.
