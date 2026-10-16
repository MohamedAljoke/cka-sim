---
id: wl-job-limits
title: Job cleanup and time limit
domain: workloads
host: cka-sim-control-plane
topics: [jobs, ttl, deadlines]
weight: 3
---
## Task

In Namespace `wl-job-limits`, create a Job `migrate` (image `busybox:1.36`, command
`sh -c 'echo migrating; sleep 5; echo done'`) that:

- is stopped and marked failed if it runs longer than `60` seconds in total
- is deleted automatically, with its Pods, `300` seconds after it finishes

Let it complete.
