---
id: nw-netpol-egress
title: The backend may only talk to the database
domain: networking
host: cka-sim-control-plane
topics: [networkpolicy, egress]
weight: 4
---
## Context

Namespace `nw-netpol-egress` runs Pod `backend` (label `app=backend`), Pod `frontend` (label
`app=frontend`), Pod `db` (label `app=db`, listening on TCP `5432` and `3306`) and Pod `cache`
(label `app=cache`, listening on TCP `6379`).

## Task

Create a NetworkPolicy `backend-egress` in Namespace `nw-netpol-egress` so that `backend` can open
connections **only** to Pods labelled `app=db` on TCP `5432`. All other egress from `backend` must
be blocked, DNS included: `backend` reaches the database by IP.

Other Pods in the Namespace must not be affected.
