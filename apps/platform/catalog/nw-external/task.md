---
id: nw-external
title: Service for an outside backend
domain: networking
host: cka-sim-control-plane
topics: [services, endpointslices, selectorless]
weight: 5
---
## Context

A legacy database API runs outside the cluster's Services, at the IP written in
`/opt/course/nw-external/backend-ip`, on port `8080`. Apps in Namespace `nw-external` should reach
it by a cluster DNS name.

## Task

In Namespace `nw-external`, create:

1. A Service `legacy-db` **without a selector**, on port `80`.
2. An EndpointSlice `legacy-db-1` that sends the Service's traffic to that IP on port `8080`.

`http://legacy-db.nw-external` must then reach the legacy API from Pod `client`.
