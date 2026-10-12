---
id: st-secret-items
title: Mount one Secret key as a private file
domain: storage
host: cka-sim-control-plane
topics: [secrets, volumes, file-permissions]
weight: 3
---
## Context

Secret `tls-bundle` in Namespace `st-secret-items` holds the keys `tls.crt`, `tls.key` and
`ca.crt`. The web server only needs its private key, at `/etc/ssl/private/server.key`.

## Task

Create Pod `web` (image `busybox:1.36`, command `sleep 1d`) that mounts **only** the key `tls.key`
of Secret `tls-bundle` as the file `/etc/ssl/private/server.key`, readable only by its owner
(mode `0400`). No other file may appear in `/etc/ssl/private`.
