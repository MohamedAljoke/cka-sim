---
id: nw-ingress-class
title: Nobody serves the Ingress without a class
domain: networking
host: cka-sim-control-plane
topics: [ingress, ingressclass, admission]
weight: 4
---
## Context

Ingress `legacy` in Namespace `nw-ingress-class` was created from
`/opt/course/nw-ingress-class/legacy.yaml`, which sets no `ingressClassName`. No controller picks it
up, so `http://legacy.nw-ingress-class.test` (controller at `10.96.80.80`) returns 404. The
cluster has one IngressClass, `nginx`, and none is marked as the default.

## Task

1. Mark IngressClass `nginx` as the cluster's default IngressClass.
2. Make Ingress `legacy` served by it **without** adding `ingressClassName` to `legacy.yaml`: create
   it again from the unchanged file.
