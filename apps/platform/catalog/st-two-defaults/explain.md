The **DefaultStorageClass** admission plugin fills in `storageClassName` for new PVCs that don't
set one. With two classes marked default, it picks the **newest**. Clusters are allowed to be in
that state (it happens during migrations), but it's confusing, so keep exactly one:

```sh
kubectl get sc                                   # (default) next to both names
kubectl annotate sc legacy storageclass.kubernetes.io/is-default-class=false --overwrite
```

Leaving the field out and setting it to `""` look alike but mean opposite things:

| PVC spec | Result |
|---|---|
| no `storageClassName` | the admission plugin sets the default class: **standard** |
| `storageClassName: ""` | explicitly **no class**: binds only to a pre-made PV that has no class either, and nothing is provisioned |

```sh
kubectl create -f pvc.yaml --dry-run=server -o jsonpath='{.spec.storageClassName}'   # see what admission does
```

A PVC created while there was no default class keeps an empty class. Since v1.28 the controller
fills it in retroactively once a default appears, unless it was set to `""` on purpose.
