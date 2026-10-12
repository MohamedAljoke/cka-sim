---
id: ar-finalizer
title: PVC stuck in Terminating
domain: architecture
host: cka-sim-control-plane
topics: [finalizers, pvc, deletion]
weight: 4
---
## Context

Someone deleted PVC `data` in Namespace `ar-finalizer` a while ago, but it is still shown as
`Terminating`. Pod `legacy-app`, which still uses it, is no longer needed.

## Task

Find out what holds the deletion and let it finish, so that PVC `data` is really gone.

Do not remove finalizers by hand.
