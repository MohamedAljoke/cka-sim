A Service balances across **every** Ready Pod that matches its selector. It doesn't care which
Deployment they belong to. Select on a label that both versions share (`app=web`) and run them as
two Deployments, and the traffic split follows the replica counts: 3 stable + 1 canary means
about 25% canary.

```sh
kubectl -n wl-canary get deploy web-stable -o yaml > canary.yaml
# rename to web-canary, replicas 1, track: canary in selector and template labels, VERSION=canary
kubectl apply -f canary.yaml
kubectl -n wl-canary get endpointslices -l kubernetes.io/service-name=web      # 4 addresses
for i in $(seq 10); do kubectl -n wl-canary exec client -- wget -qO- http://web; done
```

The `track` label must be part of each Deployment's **selector**. Otherwise the two Deployments
select each other's Pods and fight over them. Promote the canary by scaling it up and stable down,
or abort it with `scale --replicas=0`.

This only splits by Pod count. Exact percentages (say 5%) or header-based routing need weighted
routing from the Gateway API, an ingress controller or a service mesh.
