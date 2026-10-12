---
id: ar-sa-kubeconfig
title: Kubeconfig for a ServiceAccount
domain: architecture
host: cka-sim-control-plane
topics: [serviceaccounts, tokens, kubeconfig, rbac]
weight: 5
---
## Context

ServiceAccount `ci-bot` in Namespace `ar-sa-kubeconfig` may list Pods there. The CI system needs a
kubeconfig file for it.

## Task

Create a standalone kubeconfig file `/opt/course/ar-sa-kubeconfig/ci-bot.kubeconfig` that:

- contains a cluster `cka-sim` with server `https://cka-sim-control-plane:6443` and the cluster CA
  embedded
- authenticates with a token of ServiceAccount `ci-bot` that is valid for **1 hour**
- has a context `ci-bot` (Namespace `ar-sa-kubeconfig`) set as the current context

`kubectl --kubeconfig /opt/course/ar-sa-kubeconfig/ci-bot.kubeconfig get pods` must work.
