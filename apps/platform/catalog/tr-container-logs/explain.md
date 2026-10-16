`kubectl logs` reads one container's stdout and stderr, which the kubelet keeps in
`/var/log/pods/` on the node. In a Pod with more than one container you must say which one with
`-c`. Without it kubectl reads the default container, which is the first one (`proxy` here)
unless the `kubectl.kubernetes.io/default-container` annotation names another, and both
containers here log `ERROR` lines.

```sh
kubectl -n tr-container-logs get pod checkout -o jsonpath='{.spec.containers[*].name}'
kubectl -n tr-container-logs logs checkout -c app | grep ERROR > /opt/course/tr-container-logs/app-errors.log
cat /opt/course/tr-container-logs/app-errors.log
```

`grep` keeps the lines in log order. Useful neighbours: `--all-containers` (every container,
with `--prefix` to tell them apart), `--since=1h`, `--tail=20`, `-f` to follow, and `--previous`
for a crashed container.

The trap is `kubectl logs checkout | grep ERROR` without `-c`: kubectl prints only a
`Defaulted container "proxy"` note to stderr and hands you the proxy's errors.
