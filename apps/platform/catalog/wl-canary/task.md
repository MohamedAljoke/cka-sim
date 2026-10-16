---
id: wl-canary
title: Canary release with two Deployments
domain: workloads
host: cka-sim-control-plane
topics: [deployments, services, labels, releases]
weight: 4
---
## Context

Service `web` in Namespace `wl-canary` selects `app=web` and sends traffic to the 3 Pods of
Deployment `web-stable` (labels `app=web, track=stable`).

## Task

Start a canary: create Deployment `web-canary` with `1` replica so that about a quarter of the
requests to Service `web` reach it. Its Pods use the same image and command as `web-stable`, the
labels `app=web, track=canary`, and the env var `VERSION=canary`.

Do not modify Service `web` or Deployment `web-stable`.
