---
id: wl-hpa-fight
title: The pipeline and the HPA fight
domain: workloads
host: cka-sim-control-plane
topics: [hpa, autoscaling, kubectl-apply]
weight: 6
---
## Context

The team's pipeline deploys with `kubectl apply -f /opt/course/wl-hpa-fight/web.yaml` on every
commit. HPA `web` in Namespace `wl-hpa-fight` keeps Deployment `web` at 3 or more replicas, but
each deploy drops it back to 2 until the HPA scales it up again.

## Task

Make `kubectl apply -f /opt/course/wl-hpa-fight/web.yaml` stop changing the replica count of
Deployment `web`, so that only HPA `web` decides it. Right after an apply, `web` must still have the
replica count the HPA chose.

The pipeline keeps using that command and that file. Do not delete Deployment `web` or HPA
`web`.
