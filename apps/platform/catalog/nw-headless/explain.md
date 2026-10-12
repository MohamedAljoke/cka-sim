A **headless** Service has `clusterIP: None`. kube-proxy ignores it, so there is no virtual IP
and no load balancing. CoreDNS answers its name with an **A record per Ready Pod** instead. The
client gets all addresses and chooses itself, which is what databases, clustered apps and
gRPC clients with their own balancing want.

```sh
kubectl -n nw-headless create service clusterip db --tcp=80:80 --clusterip=None
kubectl -n nw-headless exec client -- nslookup db.nw-headless.svc.cluster.local
#   Name: db.nw-headless.svc.cluster.local  Address: 10.244.1.7
#   Name: db.nw-headless.svc.cluster.local  Address: 10.244.2.4  ...
```

Compare with a normal Service, whose name returns only the ClusterIP. Combined with a StatefulSet
(`serviceName: db`), each Pod also gets its own stable name, `db-0.db.nw-headless.svc.cluster.local`.
That's why StatefulSets need a headless Service.

Only Ready Pods are returned, unless the Service sets `publishNotReadyAddresses: true`. Peers that
must find each other while starting up use that.
