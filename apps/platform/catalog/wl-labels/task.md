---
id: wl-labels
title: Find and change Pods by label
domain: workloads
host: cka-sim-control-plane
topics: [labels, selectors, annotations]
weight: 3
---
## Context

Namespace `wl-labels` runs several Pods with `app`, `tier` and `env` labels.

## Task

1. Write the names of the Pods labelled `tier=backend` **and** `env=prod`, one per line, to
   `/opt/course/wl-labels/backend-prod.txt`.
2. Add the label `team=payments` to every Pod labelled `app=web`, and to no other Pod.
3. Remove the `env` label from Pod `web-dev`.
4. Annotate Pod `api` with `owner=platform`.
