---
id: tr-scheduler
title: Pods stuck in Pending
domain: troubleshooting
host: cka-sim-control-plane
topics: [kube-scheduler, static-pods]
weight: 7
---
## Context

New Pods in the cluster stay `Pending` forever. Pod `probe` in Namespace `sched-check` is one of
them.

## Task

Fix the cluster so that Pod `probe` gets scheduled and reaches `Running`.

Do not modify, delete or recreate Pod `probe`.
