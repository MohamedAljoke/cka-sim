Two mount options shape what a container sees of a volume:

- **`subPath`** mounts one sub-directory (or file) of the volume instead of its root. The kubelet
  creates the directory if it's missing. Several containers can share one volume, each in its own
  corner.
- **`readOnly: true`** on the volumeMount makes that mount read-only for the container, whatever
  the volume itself allows.

```yaml
# writer
volumeMounts:
- {name: shared, mountPath: /logs, subPath: writer-logs}
# auditor
volumeMounts:
- {name: shared, mountPath: /audit, readOnly: true}
```

```sh
kubectl -n st-readonly-share exec auditor -- ls /audit                    # writer-logs
kubectl -n st-readonly-share exec auditor -- touch /audit/x               # Read-only file system
```

Two Pods can mount the same `ReadWriteOnce` PVC because they run on the **same node**: RWO limits
nodes, not Pods. local-path volumes are pinned to one node anyway, so both Pods are scheduled
there. On storage where Pods may land on different nodes, use `ReadOnlyMany` or `ReadWriteMany`, or
`ReadWriteOncePod` to forbid exactly this kind of sharing.
