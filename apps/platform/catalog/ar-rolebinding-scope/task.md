---
id: ar-rolebinding-scope
title: A ClusterRole for one namespace
domain: architecture
host: cka-sim-control-plane
topics: [rbac, clusterroles, rolebindings, serviceaccounts]
weight: 4
---
## Context

ClusterRole `pod-reader` already exists. ServiceAccount `ci` lives in Namespace `ar-scope-dev`.

## Task

Let ServiceAccount `ci` read Pods with the rules of ClusterRole `pod-reader`, but **only in
Namespace `ar-scope-dev`**. In every other namespace, such as `ar-scope-prod`, it must not read
Pods.

Don't create any new Role or ClusterRole.
