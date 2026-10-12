The scheduler never looks at real usage. It places Pods by their **requests**, and a node is
"full" when the sum of requests reaches its Allocatable. `kubectl describe node` adds them up for
you, no metrics-server needed:

```sh
kubectl describe node cka-sim-worker | grep -A8 'Allocated resources'
#   Resource  Requests      Limits
#   cpu       1600m (13%)   100m (0%)
#   memory    50Mi (0%)     50Mi (0%)
kubectl describe nodes | grep -E '^Name:|^  cpu '
```

Requests and usage can be far apart. A Pod requesting 1500m that sleeps uses almost nothing, yet
blocks that CPU for every other Pod. A Pod with no request (BestEffort) may use a lot while
counting as zero. `kubectl top` (metrics-server) shows usage. `describe node` shows what the
scheduler sees.

When new Pods stay Pending with `Insufficient cpu` while `top` shows idle nodes, the requests are
the cause.
