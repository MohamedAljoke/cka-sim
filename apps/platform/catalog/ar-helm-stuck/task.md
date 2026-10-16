---
id: ar-helm-stuck
title: A release stuck in pending-upgrade
domain: architecture
host: cka-sim-control-plane
topics: [helm, releases, rollback]
weight: 5
---
## Context

An upgrade of release `billing` in Namespace `ar-helm-stuck` was killed halfway. Every new upgrade
now fails with `another operation (install/upgrade/rollback) is in progress`.

## Task

Bring release `billing` back to the last revision that was deployed successfully, so that its
status is `deployed` and Deployment `billing-web` runs again.
