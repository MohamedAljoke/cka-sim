---
id: ar-clusterrole
title: Read-only ClusterRole for auditors
domain: architecture
host: cka-sim-control-plane
topics: [rbac, clusterroles, clusterrolebindings]
weight: 4
---
## Task

The audit team needs to read cluster-wide storage and node information, and nothing else.

1. Create a ClusterRole `node-viewer` that can only `get` and `list` Nodes and PersistentVolumes.
2. Bind it with a ClusterRoleBinding `node-viewer` to the group `auditors`.
