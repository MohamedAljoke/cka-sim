---
id: wl-secret
title: Secret as env vars and a volume
domain: workloads
host: cka-sim-control-plane
topics: [secrets, env, volumes]
weight: 4
---
## Task

In Namespace `wl-secret`:

1. Create a Secret `db-creds` with `username=admin` and `password=S3cr3t!`.
2. Create Pod `db-client` (image `busybox:1.36`, command `sleep 1d`) that:
   - gets every key of `db-creds` as env vars (`username`, `password`) in one go
   - mounts `db-creds` **read-only** at `/etc/db-creds`
