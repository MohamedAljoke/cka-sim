---
id: tr-native-sidecar
title: Native sidecar for a log file
domain: troubleshooting
host: cka-sim-control-plane
topics: [sidecars, init-containers, logs]
weight: 4
---
## Context

Pod `report` in Namespace `tr-native-sidecar` is created from `/opt/course/tr-native-sidecar/report.yaml`.
Its app writes to `/var/log/report/report.log` instead of stdout.

## Task

Add a **native sidecar** container `log-shipper` (image `busybox:1.36`) to the Pod: a container
that starts **before** the app and keeps running beside it, streaming the log file to its own
stdout. Use the Kubernetes sidecar feature, not a second regular container.

Recreate Pod `report` from the changed file. `kubectl -n tr-native-sidecar logs report -c
log-shipper` must show the app's lines.
