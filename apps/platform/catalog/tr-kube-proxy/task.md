---
id: tr-kube-proxy
title: New Services don't route
domain: troubleshooting
host: cka-sim-control-plane
topics: [kube-proxy, services, daemonsets, iptables]
weight: 6
order: last
---
## Context

Older Services in the cluster still work, but Service `web` in Namespace `tr-kube-proxy`, created
today, doesn't answer: `wget http://web.tr-kube-proxy` from Pod `client` times out. Its Pods are
Ready and it has endpoints.

## Task

Find out why and fix the cluster, so that `http://web.tr-kube-proxy` serves the nginx welcome page
from Pod `client`.

Do not modify Service `web` or Deployment `web`.
