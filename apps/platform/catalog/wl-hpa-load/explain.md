A busybox loop is the classic load generator from the Kubernetes HPA walkthrough:

```sh
kubectl -n wl-hpa-load run load --image=busybox:1.36 -- \
  sh -c 'while true; do wget -q -O- http://web >/dev/null; done'
kubectl -n wl-hpa-load get hpa web --watch
```

`web` requests only `20m` CPU, so a few dozen requests per second push its utilization far above
the 50% target. Within a minute or two the HPA raises the replicas: metrics-server measures every
15 seconds, and the HPA acts on the next measurement.

Stop the load with `kubectl -n wl-hpa-load delete pod load`. The replicas stay up for a while
because of the HPA's scale-down stabilization window:

```sh
kubectl -n wl-hpa-load get hpa web -o jsonpath='{.spec.behavior.scaleDown.stabilizationWindowSeconds}'
```

This one sets `420`; without a `behavior` section the default is `300` seconds. During the window
the HPA uses the highest recommendation it made, so short dips in traffic don't remove Pods that
are needed again a minute later.
