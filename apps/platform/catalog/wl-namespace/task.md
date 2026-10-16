---
id: wl-namespace
title: Work across namespaces
domain: workloads
host: cka-sim-control-plane
topics: [namespaces, labels]
weight: 3
---
## Task

1. Create the Namespace `galaxy` with the label `env=training`.
2. In `galaxy`, run a Pod `probe` and a Deployment `web` with **2** replicas, both with image
   `nginx:1.27`. All their Pods must be Ready.
3. A Pod named `hermes` runs in one of the Namespaces `planet-a`, `planet-b` or `planet-c`. Write the
   name of its Namespace to `/opt/course/wl-namespace/hermes.txt`.
