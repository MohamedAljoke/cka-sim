The HPA's events and conditions say what's missing:

```sh
kubectl -n wl-hpa-unknown describe hpa shop
```

```text
ScalingActive  False  FailedGetResourceMetric  ... missing request for cpu in container web of Pod shop-...
```

A `Utilization` target is a percentage of the CPU **request**. With no request there is nothing to
divide by, so the HPA shows `<unknown>` and doesn't scale. Every container of the Pod needs one,
not only the main one: the sidecar `metrics` causes the same error after you fix `web`.

```sh
kubectl -n wl-hpa-unknown set resources deployment shop --requests=cpu=50m
```

Without `-c`, `set resources` changes every container. The Deployment rolls out new Pods, and the
HPA reads their usage on its next pass, within about 30 seconds.

The other causes of `<unknown>`: metrics-server isn't running (`kubectl top pods` fails too), or
the Pods are brand new and have no measurement yet.
