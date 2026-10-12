---
id: wl-image-digest
title: Pin an image by digest
domain: workloads
host: cka-sim-control-plane
topics: [images, deployments, supply-chain]
weight: 4
---
## Context

Deployment `web` in Namespace `wl-image-digest` runs `cka-sim-registry:5000/team/web:1.1`. Tags in
that registry get moved now and then, so a restart could run a different image than today.

## Task

1. Change Deployment `web` to reference the image it runs **right now** by its digest
   (`cka-sim-registry:5000/team/web@sha256:…`), not by a tag.
2. Set its container's `imagePullPolicy` to `IfNotPresent`.
3. Write the digest, in the form `sha256:…`, to `/opt/course/wl-image-digest/digest.txt`.
