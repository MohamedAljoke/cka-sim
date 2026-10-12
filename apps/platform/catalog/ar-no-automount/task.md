---
id: ar-no-automount
title: Pods without an API token
domain: architecture
host: cka-sim-control-plane
topics: [serviceaccounts, tokens, security]
weight: 3
---
## Context

The batch jobs in Namespace `ar-no-automount` never talk to the Kubernetes API, but every Pod still
gets a ServiceAccount token mounted.

## Task

1. Create a ServiceAccount `batch-runner` that does **not** mount its API token into Pods by
   default.
2. Create Pod `runner` (image `busybox:1.36`, command `sleep 1d`) that uses `batch-runner`, and also
   opts out of the token in its own spec.

Pod `runner` must have no `/var/run/secrets/kubernetes.io/serviceaccount` directory.
