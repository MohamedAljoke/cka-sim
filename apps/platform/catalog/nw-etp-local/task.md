---
id: nw-etp-local
title: NodePort answers on one node only
domain: networking
host: cka-sim-control-plane
topics: [services, nodeport, externaltrafficpolicy, kube-proxy]
weight: 5
---
## Context

NodePort Service `web` in Namespace `nw-etp-local` (nodePort `30090`) answers on some nodes but
times out on others. Deployment `web` runs 1 replica.

## Task

1. Test `http://<node IP>:30090` from `cka-sim-control-plane` against the two worker nodes. Write
   the name of each worker that answers right now, one per line, to
   `/opt/course/nw-etp-local/answering.txt`.
2. Change Service `web` so that port `30090` answers on **every** node. Keep 1 replica and the
   nodePort.
