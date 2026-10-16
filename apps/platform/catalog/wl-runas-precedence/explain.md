Several settings exist at both levels: `runAsUser`, `runAsGroup`, `runAsNonRoot`,
`seccompProfile`, `seLinuxOptions`. The **container** value wins where it's set, and the **Pod**
value is the default for every container that doesn't set its own:

```yaml
spec:
  securityContext:
    runAsUser: 1000        # default for all containers
  containers:
  - name: app              # -> 1000
  - name: sidecar
    securityContext:
      runAsUser: 2000      # -> 2000, overrides the Pod value
```

```sh
kubectl -n wl-runas-precedence exec two -c app -- id
kubectl -n wl-runas-precedence exec two -c sidecar -- id
```

Without either, the container runs as the image's `USER`, which is root for many images, busybox
included. `fsGroup` and `supplementalGroups` exist only at Pod level (they're about shared
volumes), while `capabilities`, `privileged` and `readOnlyRootFilesystem` exist only per container.
