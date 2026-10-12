---
id: wl-pod-failure-policy
title: Fail a Job fast on a fatal exit code
domain: workloads
host: cka-sim-control-plane
topics: [jobs, pod-failure-policy, backoff]
weight: 4
---
## Context

The importer exits with code `42` when its input file is invalid. Retrying never helps, but other
exit codes are temporary errors worth retrying.

## Task

In Namespace `wl-pod-failure-policy`, create a Job `importer` (image `busybox:1.36`, container name
`importer`, command `sh -c 'echo invalid input; exit 42'`) with `backoffLimit: 6` that:

- **fails at once**, without retrying, when container `importer` exits with code `42`
- retries other failures as usual

Let it run.
