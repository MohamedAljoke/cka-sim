---
id: nw-gateway-rewrite
title: Strip a path prefix with URLRewrite
domain: networking
host: cka-sim-control-plane
topics: [gateway-api, httproute, urlrewrite]
weight: 4
---
## Context

HTTPRoute `legacy` in Namespace `nw-gateway-rewrite` (Gateway `gw`) sends requests for
`legacy.nw-gateway-rewrite.test` whose path starts with `/legacy` to Service `app`. But `app` knows
nothing about `/legacy`: `/legacy/items` should reach it as `/items`. Its answer shows the path it
got.

## Task

Add a filter to HTTPRoute `legacy` that removes the `/legacy` prefix before the request reaches
`app`, so `/legacy/items` arrives as `/items` and `/legacy` as `/`.
