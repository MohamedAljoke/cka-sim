A **native sidecar** (GA in v1.33) is an init container with `restartPolicy: Always`:

```yaml
spec:
  initContainers:
  - name: log-shipper
    image: busybox:1.36
    restartPolicy: Always          # this one field makes it a sidecar
    command: [sh, -c, "tail -n+1 -F /var/log/report/report.log"]
    volumeMounts:
    - {name: logs, mountPath: /var/log/report}
  containers:
  - name: app
    ...
```

Compared with a plain second container, the kubelet:

- **starts it before** the app containers, and doesn't wait for it to exit (only for its startup
  probe, if any), so the app never logs before the shipper is ready
- **stops it after** the app containers on shutdown, so the last lines are still shipped
- doesn't count it when deciding whether a Job's Pod has finished, so Jobs with sidecars complete
  normally

```sh
kubectl replace --force -f report.yaml          # containers can't be added to a running Pod
kubectl -n tr-native-sidecar logs report -c log-shipper
```

Most of a Pod's spec is immutable. To change its containers, delete and recreate it.
`replace --force` does both in one step.
