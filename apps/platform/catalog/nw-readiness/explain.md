A Service only sends traffic to Pods that are **Ready**. The EndpointSlice controller lists every
Pod matching the selector, but marks an endpoint `ready: false` while its readiness probe fails,
and kube-proxy skips those. So a Pod can be `Running` and still get no traffic. `kubectl get pods`
shows it as `0/1`.

```sh
kubectl -n nw-readiness get pods                          # 0/1 Running
kubectl -n nw-readiness get endpointslices -o wide        # no ready addresses
kubectl -n nw-readiness describe pod -l app=web | grep -i readiness
#   Readiness probe failed: HTTP probe failed with statuscode: 404
kubectl -n nw-readiness edit deployment web               # httpGet.path: /
```

nginx has no `/healthz` page, so the probe gets a 404 and never passes. Point it at a path the
app really serves. Removing the probe would make the Service work too, but then Pods get traffic
before they're ready to serve, and the task forbids it.

Readiness differs from liveness: a failing readiness probe only removes the Pod from Service
endpoints, while a failing liveness probe restarts the container.
