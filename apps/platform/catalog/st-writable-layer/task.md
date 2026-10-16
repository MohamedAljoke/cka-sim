---
id: st-writable-layer
title: Data lost on container restart
domain: storage
host: cka-sim-control-plane
topics: [volumes, emptydir, containers]
weight: 3
---
## Context

Pod `scratch` in Namespace `st-writable-layer` (created from
`/opt/course/st-writable-layer/scratch.yaml`) writes `/cache/marker` the first time it starts.
After every container restart the file is gone and gets written again, because it only lives in
the container's own writable layer.

## Task

1. Change `scratch.yaml` so that `/cache` survives **container restarts** (an `emptyDir`), and
   recreate Pod `scratch` from it.
2. Restart its container once (for example `kubectl exec scratch -- kill 1`) and make sure
   `/cache/marker` from the first start is still there.
