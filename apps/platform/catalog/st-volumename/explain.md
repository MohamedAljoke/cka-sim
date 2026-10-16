When several PVs match a claim, the binder picks one of them, the smallest that fits, and you
can't tell which in advance. Two ways to pin the pair:

- **On the PVC:** `spec.volumeName: pv-orders-b` binds to exactly that PV, if it matches the class,
  modes and size, and is free.
- **On the PV:** `spec.claimRef: {namespace: st-volumename, name: orders}` reserves the PV for that
  claim before the claim even exists, so no other PVC can grab it in the meantime.

```yaml
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: orders
  volumeName: pv-orders-b
  resources:
    requests:
      storage: 1Gi
```

```sh
kubectl get pv                                   # pv-orders-b Bound to st-volumename/orders, a still Available
```

The trap: `volumeName` doesn't bypass the matching rules. A claim for `2Gi` or another class stays
`Pending` even with the right name. Use both fields together when a PV holds data for one specific
app, as with migrated or restored volumes.
