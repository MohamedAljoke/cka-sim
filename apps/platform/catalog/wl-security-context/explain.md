`securityContext` exists at two levels. **Pod-level** fields (`runAsUser`, `runAsGroup`,
`fsGroup`, `runAsNonRoot`, `seccompProfile`) apply to every container. **Container-level** fields
override them and add more: `readOnlyRootFilesystem`, `allowPrivilegeEscalation`, `capabilities`
and `privileged` exist only per container.

```yaml
spec:
  securityContext:
    runAsUser: 1000
    runAsGroup: 3000
  containers:
  - name: secure
    securityContext:
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
      capabilities:
        drop: [ALL]
    volumeMounts:
    - {name: tmp, mountPath: /tmp}     # a read-only root still needs somewhere to write
  volumes:
  - {name: tmp, emptyDir: {}}
```

```sh
kubectl -n wl-security-context exec secure -- id                 # uid=1000 gid=3000
kubectl -n wl-security-context exec secure -- touch /x           # Read-only file system
kubectl -n wl-security-context exec secure -- grep Cap /proc/1/status
```

The trap is putting `capabilities` or `readOnlyRootFilesystem` at the Pod level: the API server
rejects unknown fields there, or `kubectl apply` drops them with a warning. Many apps also need a
writable `/tmp` or cache directory, and that is what the `emptyDir` is for.
