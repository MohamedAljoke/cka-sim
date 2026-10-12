---
id: ar-merge-kubeconfig
title: Merge kubeconfigs and switch context
domain: architecture
host: cka-sim-control-plane
topics: [kubeconfig, contexts, kubectx, kubens]
weight: 4
---
## Context

`/opt/course/ar-merge-kubeconfig/` holds `team-a.kubeconfig` and `team-b.kubeconfig`, each with one
context.

## Task

1. Merge both into one self-contained file `/opt/course/ar-merge-kubeconfig/merged.kubeconfig`
   that holds both contexts with their certificates embedded, not as file paths.
2. In the merged file, make `team-b` the current context and set its default namespace to
   `ar-merge-b`, like `kubectx team-b` and `kubens ar-merge-b` would.
