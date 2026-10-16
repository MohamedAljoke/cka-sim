---
id: st-emptydir
title: Share files between containers
domain: storage
host: cka-sim-control-plane
topics: [emptydir, volumes, multi-container]
weight: 4
---
## Task

In Namespace `st-emptydir`:

1. Create Pod `share` with two containers (image `busybox:1.36`) that share an `emptyDir` volume
   mounted at `/data` in both:
   - `writer` runs `sh -c 'while true; do date >> /data/out.txt; sleep 2; done'`
   - `reader` runs `sh -c 'sleep 1d'`
2. Create Pod `share-mem`, the same as `share`, but with its `emptyDir` kept in memory and limited
   to `64Mi`.

The `reader` containers must see the file the `writer` containers write.
