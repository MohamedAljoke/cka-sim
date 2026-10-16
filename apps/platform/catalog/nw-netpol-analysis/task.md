---
id: nw-netpol-analysis
title: Which connections do the policies allow
domain: networking
host: cka-sim-control-plane
topics: [networkpolicy, analysis]
weight: 4
---
## Context

Namespace `nw-netpol-analysis` runs Pods labelled `app=web`, `app=api` and `app=db`, and
NetworkPolicies `default-deny`, `api-from-web` and `db-from-api`. Read them carefully and reason it out, as
you would on paper. The Pods run no servers, so testing won't help.

## Task

For each connection below, write whether the policies **allow** or **deny** it to
`/opt/course/nw-netpol-analysis/answers.txt`, one line each as `<number>: allowed` or
`<number>: denied`.

1. `web` → `api` on TCP `8080`
2. `web` → `db` on TCP `5432`
3. `api` → `db` on TCP `5432`
4. `api` → `web` on TCP `80`
5. a Pod labelled `app=web` in Namespace `nw-netpol-other` → `api` on TCP `8080`
6. `api` → `db` on TCP `3306`
