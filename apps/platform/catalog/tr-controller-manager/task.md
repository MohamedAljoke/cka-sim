---
id: tr-controller-manager
title: Deployments create no Pods
domain: troubleshooting
host: cka-sim-control-plane
topics: [kube-controller-manager, static-pods, deployments]
weight: 6
order: last
---
## Context

Deployment `web` in Namespace `tr-controller-manager` was created, but it has no ReplicaSet and no
Pods. Other Deployments don't scale any more either.

## Task

Find out why and fix the cluster, so that Deployment `web` runs 2 Ready Pods.

Do not modify, delete or recreate Deployment `web`.
