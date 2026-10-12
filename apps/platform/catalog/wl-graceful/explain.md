When a Pod is deleted, two things start **at the same time**: the endpoint controllers remove it
from Services (and kube-proxy and load balancers catch up a little later), and the kubelet starts
stopping its containers. If the app exits at once, requests still in flight, or still routed to
the old IP, fail.

The fix is to delay the shutdown with a **preStop hook**. The kubelet runs it first, waits for it,
and only then sends `SIGTERM`:

```yaml
spec:
  terminationGracePeriodSeconds: 45     # total budget: preStop + SIGTERM handling, then SIGKILL
  containers:
  - name: nginx
    lifecycle:
      preStop:
        exec:
          command: [sleep, "10"]        # or, on new clusters: preStop: {sleep: {seconds: 10}}
```

`terminationGracePeriodSeconds` (default 30) covers the hook **and** the app's own shutdown. When
it runs out, the container is killed with `SIGKILL` no matter what. So keep it larger than the
preStop sleep plus the app's drain time.

`postStart` is the other lifecycle hook. It runs right after the container starts, with no
guarantee that it runs before the entrypoint.
