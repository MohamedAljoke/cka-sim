---
id: wl-runas-precedence
title: Pod vs container securityContext
domain: workloads
host: cka-sim-control-plane
topics: [security-context, users, pods]
weight: 3
---
## Task

In Namespace `wl-runas-precedence`, create Pod `two` (image `busybox:1.36`, command `sleep 1d` in
both containers) with:

- a **Pod-level** `runAsUser: 1000`
- container `app`, with no user setting of its own
- container `sidecar`, with its own `runAsUser: 2000`

Then write the UID each container really runs as to `/opt/course/wl-runas-precedence/uids.txt`:

```
app: <uid>
sidecar: <uid>
```
