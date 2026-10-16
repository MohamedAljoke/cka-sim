Volume files have owners and modes like any file. A Secret volume with `defaultMode: 0440` is
readable by its owner (root) and its **group**, and by nobody else. A container running as user
1000 in group 1000 is neither, so it gets `Permission denied`.

`securityContext.fsGroup` (Pod level) fixes that without running as root. The kubelet:

1. sets the group of the volume's files to `fsGroup` (and makes new files inherit it), and
2. adds `fsGroup` to the supplementary groups of every container process.

```yaml
spec:
  securityContext:
    runAsUser: 1000
    runAsGroup: 1000
    fsGroup: 2000
```

```sh
kubectl -n st-fsgroup exec reader -- id                       # groups=1000,2000
kubectl -n st-fsgroup exec reader -- ls -lL /creds            # -r--r----- root 2000 password
```

fsGroup applies to volumes the kubelet manages (emptyDir, Secret, ConfigMap, projected, and most
CSI volumes). It doesn't apply to hostPath. On large volumes the recursive ownership change can
slow Pod start; `fsGroupChangePolicy: OnRootMismatch` skips it when the top directory is already
right.
