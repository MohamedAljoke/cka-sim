---
id: wl-psa
title: Enforce the restricted Pod Security Standard
domain: workloads
host: cka-sim-control-plane
topics: [pod-security-admission, security-context, namespaces]
weight: 5
---
## Context

Deployment `web` in Namespace `wl-psa` runs as root with default security settings.

## Task

1. Make the Namespace `wl-psa` **enforce** the `restricted` Pod Security Standard (latest
   version).
2. Fix Deployment `web` so that its Pods are admitted under that standard and run with 2 Ready
   replicas. Keep its image and command.
