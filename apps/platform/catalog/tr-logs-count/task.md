---
id: tr-logs-count
title: Count errors across a Deployment
domain: troubleshooting
host: cka-sim-control-plane
topics: [logs, label-selectors, kubectl]
weight: 3
---
## Context

Deployment `worker` in Namespace `tr-logs-count` runs 3 Pods. Some of them log `ERROR` lines.

## Task

Write one line per Pod of Deployment `worker` to `/opt/course/tr-logs-count/errors.txt`, in the
form `<pod name>: <number of log lines containing ERROR>`.
