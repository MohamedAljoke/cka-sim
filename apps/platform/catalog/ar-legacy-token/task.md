---
id: ar-legacy-token
title: Long-lived ServiceAccount token
domain: architecture
host: cka-sim-control-plane
topics: [serviceaccounts, tokens, secrets]
weight: 4
---
## Context

An old external CI system can't renew tokens. It needs a long-lived token for ServiceAccount
`deployer` in Namespace `ar-legacy-token`.

## Task

1. Create a Secret `deployer-token` of type `kubernetes.io/service-account-token` for ServiceAccount
   `deployer`, so that Kubernetes fills in a token for it.
2. Write the decoded token to `/opt/course/ar-legacy-token/token`.
