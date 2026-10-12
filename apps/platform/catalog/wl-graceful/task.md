---
id: wl-graceful
title: Graceful shutdown on rollouts
domain: workloads
host: cka-sim-control-plane
topics: [lifecycle-hooks, prestop, termination]
weight: 3
---
## Context

Every rollout of Deployment `web` in Namespace `wl-graceful` causes failed requests: Pods stop
before the load balancers have removed them.

## Task

Change Deployment `web` so that each container, when asked to stop, first runs `sleep 10` in a
`preStop` hook, and Pods get `45` seconds in total to shut down. The Deployment must roll out
fully.
