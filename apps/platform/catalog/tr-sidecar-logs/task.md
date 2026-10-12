---
id: tr-sidecar-logs
title: Stream a log file with a sidecar
domain: troubleshooting
host: cka-sim-control-plane
topics: [logs, sidecars, emptydir, multi-container]
weight: 4
---
## Context

The app in Deployment `orders` (Namespace `tr-sidecar-logs`) writes its log to the file
`/var/log/orders/orders.log` instead of stdout, so `kubectl logs` shows nothing useful.

## Task

Add a container `log-streamer` (image `busybox:1.36`) to Deployment `orders` that streams that
file to its own stdout, so that `kubectl -n tr-sidecar-logs logs deploy/orders -c log-streamer`
shows the app's log lines.

Do not change the existing container `app`.
