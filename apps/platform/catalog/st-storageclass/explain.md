A **StorageClass** tells a provisioner how to create volumes on demand. A PVC that names the class
gets a PV made for it:

| Field | Meaning |
|---|---|
| `provisioner` | which driver creates the volume (`rancher.io/local-path`, `ebs.csi.aws.com`...) |
| `reclaimPolicy` | what happens to the PV when its PVC is deleted: `Delete` (default) or `Retain` |
| `volumeBindingMode` | `Immediate` (provision at once) or `WaitForFirstConsumer` (wait for a Pod, so the volume lands where the Pod can run) |
| `allowVolumeExpansion` | whether PVCs may grow |

The **default** class is the one with the annotation
`storageclass.kubernetes.io/is-default-class: "true"`. The DefaultStorageClass admission plugin
writes it into every new PVC that has no `storageClassName`.

```sh
kubectl get storageclass                                   # (default) next to the name
kubectl annotate sc standard storageclass.kubernetes.io/is-default-class=false --overwrite
kubectl annotate sc fast-local storageclass.kubernetes.io/is-default-class=true --overwrite
```

With two defaults, Kubernetes picks the newest one, which is confusing. Always unmark the old
one. Note that `storageClassName: ""` explicitly means "no class: bind only to a pre-made PV",
which is not the same as leaving the field out.
