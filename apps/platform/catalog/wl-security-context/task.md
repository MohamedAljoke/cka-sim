---
id: wl-security-context
title: Lock down a Pod with securityContext
domain: workloads
host: cka-sim-control-plane
topics: [security-context, capabilities, users]
weight: 4
---
## Task

In Namespace `wl-security-context`, create Pod `secure` (image `busybox:1.36`, command `sleep 1d`)
that:

- runs as user `1000` and group `3000`
- has a read-only root filesystem, with a writable `emptyDir` mounted at `/tmp`
- drops **all** Linux capabilities
- can't gain more privileges than its parent process (`allowPrivilegeEscalation`)
