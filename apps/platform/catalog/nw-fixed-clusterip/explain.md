Normally the API server picks a free ClusterIP from the Service CIDR (`--service-cluster-ip-range`,
`10.96.0.0/16` here). You may choose one yourself, as long as it's inside that range and not
taken:

```sh
kubectl -n nw-fixed-clusterip create service clusterip cache --tcp=80:80 --clusterip=10.96.200.200
# or in YAML: spec.clusterIP: 10.96.200.200
kubectl -n nw-fixed-clusterip create service clusterip cache-outside --tcp=80:80 --clusterip=10.200.0.10
#   The Service "cache-outside" is invalid: spec.clusterIPs: Invalid value: ["10.200.0.10"]:
#   failed to allocate IP 10.200.0.10: the provided IP (10.200.0.10) is not in the valid range...
```

`create service clusterip` selects `app=<name>`, which matches the Deployment's Pods here. A
ClusterIP is immutable: changing it means deleting and recreating the Service.

Use fixed IPs sparingly: for DNS servers (`kube-dns` is always `10.96.0.10`, which kubelets are
configured with) or legacy clients. Everything else should use the DNS name. Kubernetes keeps the
lower part of small ranges for static assignment, so dynamic allocation rarely collides with a
hand-picked address.
