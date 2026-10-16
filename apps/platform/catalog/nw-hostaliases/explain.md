The kubelet manages each Pod's `/etc/hosts`. `spec.hostAliases` adds lines to it:

```yaml
spec:
  hostAliases:
  - ip: 10.10.0.5
    hostnames: [db.legacy.local, cache.legacy.local]
```

```sh
kubectl -n nw-hostaliases exec legacy -- cat /etc/hosts
#   # Entries added by HostAliases.
#   10.10.0.5  db.legacy.local  cache.legacy.local
```

`/etc/hosts` is checked before DNS, and it only affects this one Pod. Don't edit the file inside a
running container: the kubelet owns it and rewrites it when the Pod is recreated. Since
`hostAliases` is part of the Pod spec, it can't be changed on a running Pod; recreate it, or roll the
Deployment.

When many Pods need the same names, a CoreDNS `hosts` block, or an ExternalName or selectorless
Service, is easier to maintain than copying `hostAliases` everywhere.
