Deleting an object that has **finalizers** only sets `metadata.deletionTimestamp`, so it shows as
`Terminating`. The API server removes the object once every finalizer is gone, and each finalizer
is removed by the controller that owns it, once its cleanup condition is met.

`kubernetes.io/pvc-protection` is removed only when **no Pod uses the PVC** any more. It exists so a
running app never loses its volume under its feet.

```sh
kubectl -n ar-finalizer get pvc data -o jsonpath='{.metadata.finalizers}'
kubectl -n ar-finalizer describe pvc data | grep 'Used By'
kubectl -n ar-finalizer delete pod legacy-app
kubectl -n ar-finalizer get pvc data -w                    # disappears on its own
```

The same idea applies to PVs (`kubernetes.io/pv-protection` while bound), namespaces (stuck until
everything inside is deleted) and custom resources (the operator's finalizer).

The trap is `kubectl patch ... -p '{"metadata":{"finalizers":null}}'`. It "works", but it skips the
cleanup the finalizer guards: here, the Pod would lose its storage while running. Remove a finalizer
by hand only when its controller is gone for good.
