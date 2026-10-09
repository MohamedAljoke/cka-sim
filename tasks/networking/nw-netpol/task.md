---
id: nw-netpol
title: Restrict ingress with a NetworkPolicy
domain: networking
weight: 6
cluster: cka7491
host: cka7491-control-plane
---
## Context

Namespace `payments` runs the `api` Deployment (Service `api`, port `80`) next to other Pods.
Namespace `monitoring` runs a `prober` Pod.

## Task

Create a NetworkPolicy named `api-allow` in Namespace `payments` so that Pods with label
`app=api` accept ingress traffic on TCP port `80` **only** from:

- Pods with label `app=frontend` in Namespace `payments`
- any Pod in Namespace `monitoring`

All other ingress traffic to the `api` Pods must be denied. Do not change existing Pods.
