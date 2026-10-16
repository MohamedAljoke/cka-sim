---
id: nw-externalname
title: Alias a Service with ExternalName
domain: networking
host: cka-sim-control-plane
topics: [services, externalname, dns]
weight: 3
---
## Context

The search backend moved to Namespace `nw-externalname-search`, where it runs behind Service `api`.
Apps in Namespace `nw-externalname` still call it as `http://search`.

## Task

Create Service `search` in Namespace `nw-externalname` of type `ExternalName` that points to
`api.nw-externalname-search.svc.cluster.local`, so that `http://search` keeps working from Pod
`client`.
