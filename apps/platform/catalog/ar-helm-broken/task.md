---
id: ar-helm-broken
title: Remove a broken Helm release
domain: architecture
host: cka-sim-control-plane
topics: [helm, releases, troubleshooting]
weight: 4
---
## Context

Namespace `ar-helm-broken` has several Helm releases, and one of them never installed correctly.
The charts come from the repository `http://cka-sim-charts`.

## Task

1. Find the release in Namespace `ar-helm-broken` that failed and uninstall it. Leave the other
   releases alone.
2. Write the default values of chart `web` version `1.1.0` from that repository to
   `/opt/course/ar-helm-broken/values.yaml`.
