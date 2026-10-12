---
id: ar-openssl-user
title: Sign a user certificate with openssl
domain: architecture
host: cka-sim-control-plane
topics: [certificates, openssl, kubeconfig, rbac]
weight: 6
---
## Context

Group `ops-team` may already read everything in Namespace `ar-openssl-user`. A new member, `ops`,
needs access. The Certificates API isn't an option here, but you have the cluster CA's key on this
node.

## Task

1. Create a private key `/opt/course/ar-openssl-user/ops.key` and a certificate
   `/opt/course/ar-openssl-user/ops.crt` for user `ops` in group `ops-team`, signed directly by
   the cluster CA with `openssl`, valid for `30` days.
2. Create a standalone kubeconfig `/opt/course/ar-openssl-user/ops.kubeconfig` that uses them
   (cluster `cka-sim`, server `https://cka-sim-control-plane:6443`, CA embedded), with a context
   `ops` (Namespace `ar-openssl-user`) as the current context.
