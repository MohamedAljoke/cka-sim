---
id: nw-netpol-pick
title: Apply the least-privilege policy
domain: networking
host: cka-sim-control-plane
topics: [networkpolicy, least-privilege]
weight: 4
---
## Context

Namespace `nw-netpol-pick` runs Pod `db` (label `app=db`, listening on TCP `5432` and `9000`), Pod
`backend` (label `app=backend`) and Pod `frontend` (label `app=frontend`). Namespace
`nw-netpol-pick-other` runs another Pod labelled `app=backend`.

Four candidate policies are in `/opt/course/nw-netpol-pick/`: `policy-a.yaml` to `policy-d.yaml`.

## Task

The requirement: `db` accepts traffic only from `backend` Pods in Namespace `nw-netpol-pick`, and only
on TCP `5432`.

Apply the **one** candidate that meets the requirement with the least privilege. Don't edit it, and
don't apply the others.
