---
id: ar-helm-owner
title: Which release owns this Deployment
domain: architecture
host: cka-sim-control-plane
topics: [helm, releases, annotations]
weight: 3
---
## Context

Namespace `ar-helm-owner` has several Deployments installed by Helm, and their names don't match
their releases. Deployment `checkout` keeps being changed back after manual edits.

## Task

1. Write the name of the Helm release that owns Deployment `checkout` in Namespace
   `ar-helm-owner` to `/opt/course/ar-helm-owner/release.txt`.
2. Write a command to `/opt/course/ar-helm-owner/releases.sh` that lists the Helm releases of
   **all** Namespaces.
