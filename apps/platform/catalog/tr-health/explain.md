The API server has three health endpoints. Add `?verbose` to see each check by name, and query one
check on its own by its path:

| Endpoint | Question | Used by |
|---|---|---|
| `/livez` | should the process be restarted? | the static Pod's liveness probe |
| `/readyz` | can it serve traffic? (etcd reachable, informers synced...) | load balancers, readiness probe |
| `/healthz` | older combined endpoint, deprecated | |

```sh
kubectl get --raw='/readyz?verbose'          # [+]etcd ok ... readyz check passed
kubectl get --raw='/livez/etcd'
kubectl get --raw=/api/v1/nodes/cka-sim-worker2/proxy/healthz     # the kubelet's /healthz through the API server
curl -s http://127.0.0.1:10248/healthz       # on the node itself (healthzPort in the kubelet config)
```

`componentstatuses` asked the scheduler and controller-manager on hard-coded insecure ports that
no longer exist, so it shows errors on healthy clusters. To check those two, look at their static
Pods (`kubectl -n kube-system get pods`) or call their own `/healthz` on ports `10259` and
`10257` over HTTPS on the control plane node.
