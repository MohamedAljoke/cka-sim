---
id: nw-dns-config
title: Search domains and a custom resolver
domain: networking
host: cka-sim-control-plane
topics: [dns, resolv-conf, dnspolicy, ndots]
weight: 4
---
## Context

Pod `client` in Namespace `nw-dns-config` uses the cluster's default DNS settings.

## Task

1. From Pod `client`'s `/etc/resolv.conf`, write its search domains (space-separated, as in the
   file) and its `ndots` value to `/opt/course/nw-dns-config/client.txt`:

   ```
   search: <the search domains>
   ndots: <the ndots value>
   ```

2. Create Pod `custom-dns` (image `busybox:1.36`, command `sleep 1d`) that ignores the cluster's
   DNS settings and uses only: nameserver `10.96.0.10`, search domain
   `nw-dns-config.svc.cluster.local`, and `ndots: 2`.
