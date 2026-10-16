---
id: ar-whoami
title: Who am I, and what may they do
domain: architecture
host: cka-sim-control-plane
topics: [authentication, rbac, auth-can-i, impersonation]
weight: 3
---
## Task

1. Find out who the current kubeconfig on this node authenticates as, and write it to
   `/opt/course/ar-whoami/me.txt`:

   ```
   user: <user name>
   groups: <all groups, comma-separated>
   ```

2. Write the complete list of what ServiceAccount `viewer` in Namespace `ar-whoami` may do in that
   namespace, as `kubectl auth can-i` prints it, to `/opt/course/ar-whoami/viewer.txt`.
