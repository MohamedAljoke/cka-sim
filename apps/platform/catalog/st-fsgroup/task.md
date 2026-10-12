---
id: st-fsgroup
title: Non-root container can't read its volume
domain: storage
host: cka-sim-control-plane
topics: [security-context, fsgroup, volumes, secrets]
weight: 4
---
## Context

Pod `reader` in Namespace `st-fsgroup` (from `/opt/course/st-fsgroup/reader.yaml`) runs as user
`1000` and must read its credentials from a Secret volume whose files are only readable by their
owner and group (`defaultMode: 0440`). It fails with `Permission denied`.

## Task

Fix `reader.yaml` so that the container can read `/creds/password` **while still running as user
`1000`**, by giving the Pod's volumes the group `2000`. Keep `defaultMode: 0440`. Recreate Pod
`reader` from the file.
