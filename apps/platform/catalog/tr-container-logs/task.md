---
id: tr-container-logs
title: Error lines of one container
domain: troubleshooting
host: cka-sim-control-plane
topics: [logs, multi-container, grep]
weight: 3
---
## Context

Pod `checkout` in Namespace `tr-container-logs` runs two containers. Customers report failed
orders.

## Task

Write only the log lines of container `app` that contain `ERROR` to
`/opt/course/tr-container-logs/app-errors.log`, in the order they were logged.
