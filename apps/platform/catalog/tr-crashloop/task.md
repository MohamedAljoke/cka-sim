---
id: tr-crashloop
title: Pod in CrashLoopBackOff
domain: troubleshooting
host: cka-sim-control-plane
topics: [logs, crashloopbackoff, configmaps]
weight: 5
---
## Context

The Pod of Deployment `api` in Namespace `tr-crashloop` keeps crashing and is in `CrashLoopBackOff`.

## Task

1. Write the line in which the crashed container reports its fatal error to
   `/opt/course/tr-crashloop/error.txt`.
2. Fix the cause so that `api` runs 1 Ready Pod.

Do not change the containers or volumes of Deployment `api`. A rollout restart is allowed.
