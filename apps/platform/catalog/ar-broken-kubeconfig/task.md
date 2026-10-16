---
id: ar-broken-kubeconfig
title: Fix a broken kubeconfig
domain: architecture
host: cka-sim-control-plane
topics: [kubeconfig, contexts, troubleshooting]
weight: 4
---
## Context

A colleague's kubeconfig `/opt/course/ar-broken-kubeconfig/dev.kubeconfig` doesn't work:
`kubectl --kubeconfig /opt/course/ar-broken-kubeconfig/dev.kubeconfig get nodes` fails.

## Task

Fix that file so that the command above works, with context `dev` as its current context. Don't
replace the credentials in it with other ones.
