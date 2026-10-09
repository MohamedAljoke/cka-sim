---
id: tr-scheduler
title: Pods stuck in Pending
domain: troubleshooting
weight: 7
cluster: cka3962
host: cka3962-control-plane
---
## Context

New Pods in cluster `cka3962` stay `Pending` forever. Pod `probe` in Namespace `sched-check`
is one of them.

## Task

Fix the cluster so that Pod `probe` gets scheduled and reaches `Running`.

Do not modify, delete or recreate Pod `probe`.
