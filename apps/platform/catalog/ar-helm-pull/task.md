---
id: ar-helm-pull
title: Install a chart from a local folder
domain: architecture
host: cka-sim-control-plane
topics: [helm, charts, values]
weight: 5
---
## Context

The repository `http://cka-sim-charts` serves chart `web`. The team wants their own copy with a
different default.

## Task

1. Download chart `web` version `1.1.0` and unpack it into `/opt/course/ar-helm-pull/`, so that the
   chart is in `/opt/course/ar-helm-pull/web/`.
2. In that folder, change the chart's default `replicaCount` to `4`.
3. Install the chart from that folder as release `local` in Namespace `ar-helm-local`. The
   Namespace doesn't exist yet; have Helm create it. Don't override any value on the command
   line.
