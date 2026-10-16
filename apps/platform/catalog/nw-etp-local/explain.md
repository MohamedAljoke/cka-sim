`externalTrafficPolicy` decides what kube-proxy does with traffic that arrives at a node's
NodePort (or load balancer):

| Policy | Traffic goes to | Client IP seen by the Pod |
|---|---|---|
| `Cluster` (default) | any Ready endpoint, on any node; one extra hop, with SNAT | a node's IP |
| `Local` | only endpoints **on that same node**; dropped if there are none | the real client IP |

With 1 replica on `cka-sim-worker2` and `Local`, only worker2 answers, and the other nodes drop the
connection. That's the point of `Local`: cloud load balancers health-check the nodePort
(`healthCheckNodePort`) and send traffic only to nodes that answer, and apps get the true client
address.

```sh
kubectl -n nw-etp-local get svc web -o jsonpath='{.spec.externalTrafficPolicy}'
kubectl -n nw-etp-local get pods -o wide                     # where the endpoints are
for n in $(kubectl get nodes -o jsonpath='{.items[*].status.addresses[0].address}'); do curl -s -m2 $n:30090 >/dev/null && echo ok $n; done
kubectl -n nw-etp-local patch svc web -p '{"spec":{"externalTrafficPolicy":"Cluster"}}'
```

A node calling its **own** IP is the exception: kube-proxy treats traffic that starts on the node as
local and lets it through to any endpoint. That's why you test the workers from another machine,
here the control plane.

The other fix, if the client IP matters, is to keep `Local` and run a Pod on every node: a
DaemonSet, or more replicas spread across nodes.
