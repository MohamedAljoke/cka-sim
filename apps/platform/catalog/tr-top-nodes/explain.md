```sh
kubectl top nodes --sort-by=memory
```

`--sort-by` on `kubectl top` accepts `cpu` or `memory` and sorts in descending order, so the node
with the least memory to spare comes first. The columns are CPU in cores and as a percentage of
the node's allocatable CPU, then memory in bytes and as a percentage.

Pipes through `sort` work too, but they have to deal with units: `sort -k4 -h` doesn't understand
`Mi` and `Gi`. The built-in flag saves that.

If `kubectl top` says `Metrics API not available`, metrics-server isn't running, or its APIService
`v1beta1.metrics.k8s.io` isn't `Available`; `kubectl get apiservice` shows which.
