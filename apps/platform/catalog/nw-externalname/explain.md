An **ExternalName** Service is a pure DNS alias: CoreDNS answers its name with a **CNAME** record
to `spec.externalName`. There is no ClusterIP, no selector, no endpoints and no kube-proxy rules.

```sh
kubectl -n nw-externalname create service externalname search --external-name=api.nw-externalname-search.svc.cluster.local
kubectl -n nw-externalname exec client -- nslookup search
#   search.nw-externalname.svc.cluster.local  canonical name = api.nw-externalname-search.svc.cluster.local
```

Typical uses: point a stable in-cluster name at a managed database (`mydb.abc.rds.amazonaws.com`),
or keep an old name working after moving a Service to another namespace, as here. Switching later
to a real in-cluster Service only needs a change to this one object.

Limits: it works at DNS level only, so ports aren't remapped. HTTP and TLS still see the original
host name: a server that routes by `Host`, or a TLS certificate for the target name, may reject the
request. `externalName` must be a DNS name; for an IP address, use a selectorless Service with an
EndpointSlice.
