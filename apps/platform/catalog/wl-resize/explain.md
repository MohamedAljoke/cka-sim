A Pod's resources used to be immutable: changing them meant recreating the Pod. **In-place
resize** (GA in v1.35) changes CPU and memory of running containers through the Pod's `resize`
subresource. The kubelet adjusts the container's cgroup limits live.

```sh
kubectl -n wl-resize patch pod worker --subresource resize \
  -p '{"spec":{"containers":[{"name":"worker","resources":{"requests":{"cpu":"300m"},"limits":{"cpu":"400m"}}}]}}'
kubectl -n wl-resize get pod worker -o jsonpath='{.status.containerStatuses[0].resources}'   # what's applied
kubectl -n wl-resize get pod worker -o jsonpath='{.status.conditions}'                       # PodResizePending / PodResizeInProgress
```

`spec...resources` is what you asked for, and `status.containerStatuses[].resources` is what the
node actually applied. A resize that doesn't fit on the node stays pending (`Deferred` or
`Infeasible`). Each container's `resizePolicy` decides whether a change needs a container restart
(`RestartContainer`, common for memory with some runtimes) or not (`NotRequired`, the default).

A plain `kubectl edit pod` or `apply` on resources is still rejected: only the `resize`
subresource may change them. A Pod's QoS class can't change by resizing. For Deployments, change
the template instead; that rolls out new Pods.
