---
id: tr-health
title: Check control-plane health
domain: troubleshooting
host: cka-sim-control-plane
topics: [health-checks, kube-apiserver, kubelet]
weight: 3
---
## Context

`kubectl get componentstatuses` is deprecated and no longer tells the truth about this cluster.

## Task

1. Write the API server's detailed readiness report, with every individual check, to
   `/opt/course/tr-health/readyz.txt`.
2. Write the health response of the kubelet on node `cka-sim-worker2` to
   `/opt/course/tr-health/kubelet.txt`.
